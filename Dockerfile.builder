FROM goreleaser/goreleaser:latest

RUN apk add --no-cache nodejs npm yarn git upx wget unzip bash openjdk17-jre-headless gcompat
RUN go install github.com/swaggo/swag/cmd/swag@latest

ENV ANDROID_HOME=/opt/android-sdk
ENV NDK_TOOLCHAIN=${ANDROID_HOME}/ndk/latest/toolchains/llvm/prebuilt/linux-x86_64
ENV PATH=${PATH}:${ANDROID_HOME}/cmdline-tools/latest/bin

RUN mkdir -p ${ANDROID_HOME}/cmdline-tools && \
    wget -q https://dl.google.com/android/repository/commandlinetools-linux-14742923_latest.zip -O /tmp/cmdline.zip && \
    unzip -q /tmp/cmdline.zip -d ${ANDROID_HOME}/cmdline-tools && \
    mv ${ANDROID_HOME}/cmdline-tools/cmdline-tools ${ANDROID_HOME}/cmdline-tools/latest && \
    rm /tmp/cmdline.zip

RUN yes | sdkmanager --licenses && \
    sdkmanager "ndk;27.0.12077973" && \
    ln -s ${ANDROID_HOME}/ndk/* ${ANDROID_HOME}/ndk/latest
