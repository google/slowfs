#SlowFS
SlowFS is a FUSE filesystem written in Go to simulate physical media for testing
purposes. SlowFS works by mirroring a backing directory to another directory,
and then making accesses to files in that directory take a specified amount
of time. SlowFS can model things like seek time, throughput, sequential / non-
sequential accesses, et. cetera.

For addition details, run SlowFS with the help flag:
  `slowfs --help`

This is not an official Google product.

##Basic Usage

Example invocation:
  `slowfs --backing-dir=my-backing-dir --mount-dir=my-mount-dir`

##Configuration Files

You can specify an optional configuration file listing configurations in JSON,
and then pass that as an argument.
```json
[
  {
    "Name": "fast",
    "SeekWindow": "16KiB",
    "SeekTime": "8ms",
    "ReadBytesPerSecond": "100MiB",
    "WriteBytesPerSecond": "100MiB",
    "AllocateBytesPerSecond": "4GiB",
    "RequestReorderMaxDelay": "100us",
    "FsyncStrategy": "wbc",
    "WriteStrategy": "fastwrite",
    "MetadataOpTime": "500us"
  }
]
```

Example invocation:
  ```slowfs --backing-dir=my-backing-dir --mount-dir=my-mount-dir \
    --config-file=my-config-file.json --config-name=fast```

###Overriding Values

You can also override any option through the corresponding command line flag.
For example, if you would like to change seek time:
  ```slowfs --backing-dir=my-backing-dir --mount-dir=my-mount-dir \
    --config-file=my-config-file.json --config-name=fast --seek-time=16ms```
