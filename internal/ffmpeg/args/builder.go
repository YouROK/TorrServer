package args

import (
	"fmt"
	"strconv"
	"strings"
)

// EncoderSelector подбирает конкретный кодировщик по кодеку и ускорению.
type EncoderSelector interface {
	AutoVideoEncoder(codec, hwaccel string) string
	AutoAudioEncoder(codec string) string
}

// Build собирает аргументы командной строки ffmpeg для задания.
func Build(o Options, sel EncoderSelector) ([]string, error) {
	if strings.TrimSpace(o.Input) == "" {
		return nil, fmt.Errorf("input is required")
	}
	if o.Container == "" && o.Protocol != ProtocolHLS {
		return nil, fmt.Errorf("container is required")
	}

	// Кодировщик выбирается до сборки аргументов: от него зависит,
	// нужны ли аппаратное устройство и аппаратные фильтры
	encoder, err := resolveVideoEncoder(o, sel)
	if err != nil {
		return nil, err
	}
	hw := hwFromEncoder(encoder)

	var a []string
	a = append(a, "-hide_banner", "-nostats", "-loglevel", "error")
	a = append(a, "-progress", "pipe:2")

	// Устройство ускорения создаётся до ввода, иначе его не увидит декодер
	a = append(a, buildHwDeviceArgs(hw, o.Video.HWDevice)...)

	// Аппаратное декодирование тоже задаётся до ввода
	a = append(a, buildHwDecodeArgs(hw, o.Video)...)

	// Ввод
	if o.Realtime {
		a = append(a, "-re")
	}
	if o.AnalyzeDurMs > 0 {
		a = append(a, "-analyzeduration", strconv.Itoa(o.AnalyzeDurMs*1000))
	}
	if o.ProbeSizeMB > 0 {
		a = append(a, "-probesize", strconv.Itoa(o.ProbeSizeMB<<20))
	}
	a = append(a, o.InputArgs...)

	// Позиция старта задаётся до ввода, поэтому seek выполняется быстро
	if o.Seek > 0 {
		a = append(a, "-ss", formatTimestamp(o.Seek))
	}
	a = append(a, "-i", o.Input)

	a = append(a, buildMapArgs(o)...)
	a = append(a, buildVideoArgs(o, encoder, hw)...)
	a = append(a, buildAudioArgs(o, sel)...)

	if o.Threads > 0 {
		a = append(a, "-threads", strconv.Itoa(o.Threads))
	}

	if o.Protocol == ProtocolHLS {
		a = append(a, buildHLSArgs(o)...)
	} else {
		a = append(a, buildOutputFormatArgs(o)...)
	}

	a = append(a, o.ExtraArgs...)
	a = append(a, "-y", o.Output)
	return a, nil
}

// buildMapArgs выбирает потоки, попадающие в результат.
func buildMapArgs(o Options) []string {
	var a []string

	if o.Video.Codec == CopyCodec {
		a = append(a, "-map", "0:v:0?")
	} else if idx := o.Video.StreamIndex; idx != nil {
		a = append(a, "-map", fmt.Sprintf("0:%d", *idx))
	} else {
		a = append(a, "-map", "0:v:0?")
	}

	if o.Audio.Codec == CopyCodec {
		a = append(a, "-map", "0:a:0?")
	} else if idx := o.Audio.StreamIndex; idx != nil {
		a = append(a, "-map", fmt.Sprintf("0:%d", *idx))
	} else {
		a = append(a, "-map", "0:a:0?")
	}

	// Субтитры впечатываются фильтром, поэтому в карту не попадают
	switch o.Subtitles.Mode {
	case SubtitleCopy:
		if idx := o.Subtitles.StreamIndex; idx != nil {
			a = append(a, "-map", fmt.Sprintf("0:%d", *idx))
		} else {
			a = append(a, "-map", "0:s:0?")
		}
	case SubtitleBurn:
		// Дорожка читается фильтром subtitles, отдельный поток ей не нужен
	default:
		if o.DropSubs {
			a = append(a, "-sn")
		}
	}
	return a
}

// resolveVideoEncoder определяет кодировщик для задания.
// Пустая строка означает копирование потока без перекодирования.
func resolveVideoEncoder(o Options, sel EncoderSelector) (string, error) {
	v := o.Video
	if v.Codec == "" || v.Codec == "none" || v.Codec == CopyCodec {
		return "", nil
	}

	if v.Encoder != "" {
		return v.Encoder, nil
	}
	if sel != nil {
		if enc := sel.AutoVideoEncoder(v.Codec, v.HWAccel); enc != "" {
			return enc, nil
		}
	}
	return "", fmt.Errorf("no encoder available for codec %q", v.Codec)
}

// buildVideoArgs собирает параметры видеокодирования.
func buildVideoArgs(o Options, encoder, hw string) []string {
	v := o.Video
	hasVideo := v.Codec != "" && v.Codec != "none"
	if !hasVideo {
		return []string{"-vn"}
	}

	if v.Codec == CopyCodec {
		return []string{"-c:v", "copy"}
	}

	a := []string{"-c:v", encoder}
	a = append(a, buildVideoFilters(o, hw)...)

	if v.PixFmt != "" {
		a = append(a, "-pix_fmt", v.PixFmt)
	}
	if v.Level != "" {
		a = append(a, "-level", v.Level)
	}
	if v.Profile != "" {
		a = append(a, "-profile:v", v.Profile)
	}
	if v.MaxFramerate > 0 {
		a = append(a, "-r", strconv.FormatFloat(v.MaxFramerate, 'f', -1, 64))
	}

	a = append(a, buildRateControl(encoder, v)...)
	a = append(a, buildPresetArgs(encoder, v.Preset)...)

	// Ключевые кадры выравниваются по границам сегментов, иначе плеер не сможет перематывать
	if o.Protocol == ProtocolHLS && o.HLS.SegmentLength > 0 {
		a = append(a, buildKeyframeArgs(encoder, o.HLS.SegmentLength)...)
	}
	return a
}

// buildRateControl выбирает режим управления битрейтом.
func buildRateControl(encoder string, v VideoOptions) []string {
	hw := isHardwareEncoder(encoder)

	if v.BitrateKbps > 0 {
		bitrate := strconv.Itoa(v.BitrateKbps) + "k"
		a := []string{"-b:v", bitrate}

		maxRate := v.MaxBitrateKbps
		if maxRate <= 0 {
			maxRate = v.BitrateKbps
		}
		a = append(a, "-maxrate", strconv.Itoa(maxRate)+"k", "-bufsize", strconv.Itoa(maxRate*2)+"k")

		if hw {
			a = append(a, hardwareQualityArgs(encoder, v)...)
		}
		return a
	}

	// Аппаратные кодировщики не поддерживают CRF, для них задаётся качество через qp
	if hw {
		a := hardwareQualityArgs(encoder, v)
		if len(a) > 0 {
			return a
		}
	}
	if v.CRF > 0 {
		return []string{"-crf", strconv.Itoa(v.CRF)}
	}
	return nil
}

// hardwareQualityArgs задаёт качество для аппаратных кодировщиков.
func hardwareQualityArgs(encoder string, v VideoOptions) []string {
	q := v.CRF
	if q <= 0 {
		q = 23
	}

	switch {
	case strings.HasSuffix(encoder, "_vaapi"):
		return []string{"-qp", strconv.Itoa(q)}
	case strings.HasSuffix(encoder, "_nvenc"):
		return []string{"-cq", strconv.Itoa(q)}
	case strings.HasSuffix(encoder, "_qsv"):
		return []string{"-global_quality", strconv.Itoa(q)}
	default:
		return nil
	}
}

// buildPresetArgs добавляет пресет кодирования, если он поддерживается.
func buildPresetArgs(encoder, preset string) []string {
	if preset == "" {
		return nil
	}
	switch {
	case strings.HasPrefix(encoder, "libx264"), strings.HasPrefix(encoder, "libx265"):
		return []string{"-preset", preset}
	case strings.HasSuffix(encoder, "_nvenc"):
		return []string{"-preset", nvencPreset(preset)}
	default:
		return nil
	}
}

// nvencPreset переводит общий пресет в значение, понятное NVENC.
func nvencPreset(preset string) string {
	switch preset {
	case "ultrafast", "superfast", "veryfast":
		return "p1"
	case "faster", "fast":
		return "p2"
	case "medium":
		return "p4"
	case "slow", "slower":
		return "p6"
	case "veryslow", "placebo":
		return "p7"
	default:
		return "p4"
	}
}

// buildKeyframeArgs задаёт принудительные ключевые кадры на границах сегментов.
func buildKeyframeArgs(encoder string, segmentLength int) []string {
	expr := fmt.Sprintf("expr:gte(t,n_forced*%d)", segmentLength)
	a := []string{"-force_key_frames", expr}

	if strings.HasPrefix(encoder, "libx264") || strings.HasPrefix(encoder, "libx265") {
		a = append(a, "-sc_threshold", "0")
	}
	return a
}

// buildVideoFilters собирает цепочку видеопотоковых фильтров.
// При аппаратном декодировании кадры уже лежат в памяти устройства,
// поэтому загрузка не нужна и все фильтры должны быть аппаратными.
// При аппаратном кодировании кадр сначала загружается в устройство.
func buildVideoFilters(o Options, hw string) []string {
	v := o.Video
	decodingOnDevice := hw != "" && v.HWDecode

	var filters []string

	// Деинтерлейс выполняется на устройстве, если кадры уже там,
	// иначе программно до загрузки
	if shouldDeinterlace(v) {
		if decodingOnDevice {
			filters = append(filters, hwDeinterlaceFilter(hw))
		} else {
			filters = append(filters, "yadif=deint=interlaced")
		}
	}

	if hw != "" && !decodingOnDevice {
		filters = append(filters, "hwupload")
	}

	// Тонемаппинг выполняется до масштабирования: диапазон приводится к SDR.
	// Режим auto включает его только для HDR-источника
	if v.TonemapMode.Enabled(o.SourceHDR) {
		filters = append(filters, tonemapFilter(hw, decodingOnDevice, v)...)
	}

	// Поворот применяется до масштабирования
	if v.Transpose != "" {
		filters = append(filters, transposeFilter(v.Transpose, hw))
	}

	if v.MaxWidth > 0 || v.MaxHeight > 0 {
		if hw != "" {
			filters = append(filters, hwScaleFilter(v.MaxWidth, v.MaxHeight, hw))
		} else {
			filters = append(filters, scaleFilter(v.MaxWidth, v.MaxHeight))
		}
	}

	// Впечатывание субтитров выполняется последним: кадр уже нужного размера
	if o.Subtitles.Mode == SubtitleBurn {
		filters = append(filters, burnSubtitleFilter(o.Subtitles, o.Input))
	}

	if len(filters) == 0 {
		return nil
	}
	return []string{"-vf", strings.Join(filters, ",")}
}

// burnSubtitleFilter собирает фильтр впечатывания субтитров.
// Фильтру обязательно нужно имя исходного файла: по одной дорожке он работать не умеет.
// Номер дорожки отсчитывается внутри файла, а не в общем списке потоков.
func burnSubtitleFilter(opts SubtitleOptions, input string) string {
	filter := "subtitles=f=" + escapeFilterPath(input)

	if opts.StreamIndex != nil {
		filter += ":si=" + strconv.Itoa(*opts.StreamIndex)
	}
	if opts.ForceStyle != "" {
		filter += ":force_style='" + escapeForceStyle(opts.ForceStyle) + "'"
	}
	return filter
}

// escapeFilterPath экранирует путь для использования в фильтре.
// Специальные символы фильтров экранируются обратной косой чертой.
func escapeFilterPath(path string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		":", "\\\\:",
		"'", "\\\\'",
		"[", "\\\\[",
		"]", "\\\\]",
		",", "\\\\,",
	)
	return "'" + replacer.Replace(path) + "'"
}

// escapeForceStyle экранирует стиль субтитров для фильтра.
func escapeForceStyle(style string) string {
	return strings.ReplaceAll(style, "'", "\\'")
}

// tonemapFilter собирает фильтры приведения HDR к SDR.
// Аппаратный путь выполняется на устройстве, программный разбивается
// на перевод в 10-битный формат, собственно тонемаппинг и возврат к 8 битам.
func tonemapFilter(hw string, onDevice bool, v VideoOptions) []string {
	algorithm := v.TonemapAlgorithm
	if algorithm == "" {
		algorithm = "hable"
	}

	if hw != "" {
		switch hw {
		case "vaapi":
			// Аппаратный VPP требует, чтобы кадры уже были на устройстве
			return []string{"tonemap_vaapi=format=nv12:t=bt709:m=bt709:p=bt709"}
		case "cuda":
			return []string{"tonemap_cuda=format=yuv420p:t=bt709:m=bt709:p=bt709"}
		case "qsv":
			return []string{"vpp_tonemap=1"}
		}
	}

	// Программный путь: zscale приводит диапазон, tonemapx выполняет отображение
	return []string{
		"zscale=t=linear:npl=100,format=gbrpf32le",
		"zscale=p=bt709",
		"tonemap=tonemap=" + algorithm + ":desat=0",
		"zscale=t=bt709:m=bt709:r=tv,format=yuv420p",
	}
}

// hwDeinterlaceFilter возвращает аппаратный фильтр устранения чересстрочности.
func hwDeinterlaceFilter(hw string) string {
	switch hw {
	case "cuda":
		return "yadif_cuda=deint=interlaced"
	case "qsv":
		return "deinterlace_qsv=mode=2"
	default:
		// VAAPI и остальные используют общий фильтр с явным форматом
		return "deinterlace_vaapi=rate=field:auto=1"
	}
}

// transposeFilter возвращает фильтр поворота для программного или аппаратного пути.
func transposeFilter(mode, hw string) string {
	dir := "clock"
	switch mode {
	case "90", "clock":
		dir = "clock"
	case "180", "cclock":
		dir = "cclock"
	case "270", "reversal":
		dir = "reversal"
	}

	if hw != "" {
		return "transpose_" + hw + "=dir=" + dir
	}
	return "transpose=dir=" + dir
}

// buildHwDecodeArgs включает декодирование на устройстве.
// Кадры остаются в памяти устройства, поэтому фильтры тоже должны быть аппаратными.
func buildHwDecodeArgs(hw string, v VideoOptions) []string {
	if hw == "" || !v.HWDecode {
		return nil
	}

	switch hw {
	case "vaapi":
		// Устройство уже создано, поэтому указываем его по имени
		a := []string{"-hwaccel", "vaapi", "-hwaccel_output_format", "vaapi"}
		if v.HWDevice != "" {
			a = append(a, "-hwaccel_device", v.HWDevice)
		}
		return a
	case "cuda":
		return []string{"-hwaccel", "cuda", "-hwaccel_output_format", "cuda"}
	case "qsv":
		return []string{"-hwaccel", "qsv", "-hwaccel_output_format", "qsv"}
	default:
		return nil
	}
}

// buildHwDeviceArgs создаёт устройство аппаратного ускорения по имени фильтров.
// Пустой путь означает узел рендеринга по умолчанию.
func buildHwDeviceArgs(hw, device string) []string {
	switch hw {
	case "vaapi":
		if device == "" {
			device = DefaultHWDevice
		}
		return []string{"-init_hw_device", "vaapi=va:" + device, "-filter_hw_device", "va"}
	case "cuda":
		return []string{"-init_hw_device", "cuda=cu:0", "-filter_hw_device", "cu"}
	case "qsv":
		return []string{"-init_hw_device", "qsv=qs", "-filter_hw_device", "qs"}
	default:
		return nil
	}
}

// DefaultHWDevice - узел рендеринга VAAPI по умолчанию.
const DefaultHWDevice = "/dev/dri/renderD128"

// hwFromEncoder определяет ускорение по фактически выбранному кодировщику.
// Возвращает суффикс имён аппаратных фильтров или пустую строку.
func hwFromEncoder(encoder string) string {
	switch {
	case strings.HasSuffix(encoder, "_vaapi"):
		return "vaapi"
	case strings.HasSuffix(encoder, "_nvenc"):
		return "cuda"
	case strings.HasSuffix(encoder, "_qsv"):
		return "qsv"
	default:
		return ""
	}
}

// hwScaleFilter строит фильтр масштабирования на устройстве.
func hwScaleFilter(maxWidth, maxHeight int, hw string) string {
	w := "iw"
	h := "ih"
	if maxWidth > 0 {
		w = fmt.Sprintf("min(%d,iw)", maxWidth)
	}
	if maxHeight > 0 {
		h = fmt.Sprintf("min(%d,ih)", maxHeight)
	}

	// Поверхности устройства не кратны двум автоматически, поэтому округляем сами
	return fmt.Sprintf("scale_%s=w='%s':h='%s':force_original_aspect_ratio=decrease:force_divisible_by=2:format=nv12",
		hw, w, h)
}

// scaleFilter строит фильтр масштабирования без апскейла и с сохранением пропорций.
func scaleFilter(maxWidth, maxHeight int) string {
	w := "iw"
	h := "ih"
	if maxWidth > 0 {
		w = fmt.Sprintf("min(%d,iw)", maxWidth)
	}
	if maxHeight > 0 {
		h = fmt.Sprintf("min(%d,ih)", maxHeight)
	}
	return fmt.Sprintf("scale=w='%s':h='%s':force_original_aspect_ratio=decrease:force_divisible_by=2", w, h)
}

// shouldDeinterlace решает, нужен ли фильтр устранения чересстрочности.
func shouldDeinterlace(v VideoOptions) bool {
	return v.Deinterlace == "yadif"
}

// buildAudioArgs собирает параметры аудиокодирования.
func buildAudioArgs(o Options, sel EncoderSelector) []string {
	au := o.Audio
	if au.Codec == "" || au.Codec == "none" {
		return []string{"-an"}
	}
	if au.Codec == CopyCodec {
		return []string{"-c:a", "copy"}
	}

	encoder := au.Encoder
	if encoder == "" && sel != nil {
		encoder = sel.AutoAudioEncoder(au.Codec)
	}
	if encoder == "" {
		encoder = au.Codec
	}

	a := []string{"-c:a", encoder}
	if au.BitrateKbps > 0 {
		a = append(a, "-b:a", strconv.Itoa(au.BitrateKbps)+"k")
	}
	if au.Channels > 0 {
		a = append(a, "-ac", strconv.Itoa(au.Channels))
	}
	if au.SampleRate > 0 {
		a = append(a, "-ar", strconv.Itoa(au.SampleRate))
	}
	return a
}

// buildHLSArgs собирает параметры сегментирования HLS.
func buildHLSArgs(o Options) []string {
	h := o.HLS
	if h.SegmentLength <= 0 {
		h.SegmentLength = DefaultHLSOptions().SegmentLength
	}

	segmentType := h.SegmentType
	if segmentType == "" {
		segmentType = "mpegts"
	}

	a := []string{"-f", "hls", "-hls_time", strconv.Itoa(h.SegmentLength), "-hls_segment_type", segmentType}

	if h.SegmentFilename != "" {
		a = append(a, "-hls_segment_filename", h.SegmentFilename)
	}

	// Сегменты fmp4 требуют начальный сегмент с описанием дорожек,
	// на который плейлист ссылается через EXT-X-MAP
	if segmentType == "fmp4" {
		name := h.InitFilename
		if name == "" {
			name = DefaultInitSegment
		}
		a = append(a, "-hls_fmp4_init_filename", name)
		a = append(a, "-hls_flags", "independent_segments")
	}

	switch h.PlaylistType {
	case "vod", "event":
		a = append(a, "-hls_playlist_type", h.PlaylistType, "-hls_list_size", "0")
	default:
		// Скользящее окно: плейлист короткий, старые сегменты удаляются
		listSize := h.ListSize
		if listSize <= 0 {
			listSize = 6
		}
		threshold := h.DeleteThreshold
		if threshold <= 0 {
			threshold = 1
		}
		a = append(a, "-hls_list_size", strconv.Itoa(listSize), "-hls_delete_threshold", strconv.Itoa(threshold))
		a = append(a, "-hls_flags", "delete_segments+temp_file")
	}

	if h.StartNumber > 0 {
		a = append(a, "-start_number", strconv.Itoa(h.StartNumber))
	}

	// Метки времени входного файла не сдвигаются, иначе перемотка после перезапуска сбивается
	a = append(a, "-copyts", "-avoid_negative_ts", "disabled")
	a = append(a, "-start_at_zero")
	return a
}

// buildOutputFormatArgs собирает параметры формата вывода.
func buildOutputFormatArgs(o Options) []string {
	switch normalizeContainer(o.Container) {
	case "mp4":
		return []string{"-f", "mp4", "-movflags", "frag_keyframe+empty_moov+delay_moov", "-frag_duration", "1000000"}
	case "mkv":
		return []string{"-f", "matroska"}
	case "ts":
		return []string{"-f", "mpegts"}
	case "webm":
		return []string{"-f", "webm"}
	default:
		return []string{"-f", normalizeContainer(o.Container)}
	}
}

// normalizeContainer приводит имя контейнера к имени формата ffmpeg.
func normalizeContainer(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "mkv":
		return "mkv"
	case "matroska":
		return "mkv"
	case "mp4":
		return "mp4"
	case "ts", "mpegts":
		return "ts"
	case "webm":
		return "webm"
	default:
		return strings.ToLower(strings.TrimSpace(name))
	}
}

// isHardwareEncoder определяет, является ли кодировщик аппаратным.
func isHardwareEncoder(encoder string) bool {
	for _, suffix := range []string{"_vaapi", "_nvenc", "_qsv", "_amf", "_videotoolbox", "_v4l2m2m", "_rkmpp"} {
		if strings.HasSuffix(encoder, suffix) {
			return true
		}
	}
	return false
}

// formatTimestamp переводит секунды в формат hh:mm:ss.mmm.
func formatTimestamp(seconds float64) string {
	if seconds < 0 {
		seconds = 0
	}
	total := int64(seconds)
	ms := int64((seconds - float64(total)) * 1000)
	return fmt.Sprintf("%02d:%02d:%02d.%03d", total/3600, (total%3600)/60, total%60, ms)
}
