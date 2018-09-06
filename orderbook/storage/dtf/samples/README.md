# Compression Samples

Here exists a "library" of different compressions. 
The original file size (`b46ca9f1-654a-41b5-8bac-d366423645ee--bitmex:XBTUSD.dtf`) is `20983703 bytes`, or `20.983703 Megabytes`.

Without further adeu, here's a table demonstrating the compression capabilities and performance vs. their competitors:

| Compression Algorithm | Compression Ratio | Runtime Speed | Comments |
|:----------------------|------------------:|---------------|----------|
| **`LZMA2`** | 65.959% | Fast | Accessible |
| **`ncompress`**| 47.544% | Fast | None |
| **`bzip2`** | 62.257% | Fast | Accessible |
| **`ZPAQ (lrzip)`** | 72.547% | Slower | Rarer |
| **`paq8px`** | 79.969% | Slow (1hr) | Don't use unless you happen to have time |