# Reading memory use on Linux

`systemctl status torrserver` and `memory.current` can show two to three times the memory the process actually holds, for a while after a torrent closes. It is accounting, not a leak. Compare the numbers before you chase it:

```bash
d=/sys/fs/cgroup/system.slice/torrserver.service
cat $d/memory.current
grep -E '^(anon|active_anon|anon_thp|file|sock) ' $d/memory.stat
```

and, with the server started with `--pprof` (see Server args), the Go side:

```bash
curl -s -u user:pass http://HOST:PORT/runtime/status   # memory.heap_inuse, heap_idle, heap_released, sys
```

What to look for:

- `anon` (and `RssAnon` in `/proc/<pid>/status`) is what the process holds. `heap_released` is memory Go already handed back to the OS.
- With transparent huge pages set to `always` (`cat /sys/kernel/mm/transparent_hugepage/enabled`), Go's heap sits on 2 MiB pages. When Go gives part of a page back, the process stops mapping it (RSS drops) but the kernel keeps the whole page charged to the cgroup until its deferred split runs, which normally happens only under memory pressure. The tell is `active_anon` far above `anon`.
- Measured on a Debian 13 host (kernel 6.12, THP `always`, 4 GB RAM, no `MemoryMax`): after 30 s of streaming a 193 GB torrent and dropping it, `memory.current` stayed about 430 MiB while `anon` was 135 to 170 MiB and Go's `heap_inuse` was 61 MiB. Streaming itself peaked near 540 MiB.
- Nothing is lost: with free memory on the host the kernel releases those pages when it needs them, and a unit with `MemoryMax=` is reclaimed before it is killed. It only makes graphs and `systemctl status` look high.
- If you want clean numbers anyway, set THP to `madvise` for the host (`echo madvise | sudo tee /sys/kernel/mm/transparent_hugepage/enabled`, and `transparent_hugepage=madvise` on the kernel command line to keep it). `GODEBUG=disablethp=1` in the unit did not change this in the same test (build with Go 1.27.3), so do not count on it. Changing THP affects every process on the host.
