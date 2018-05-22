package dtf

// CompressDaemon : We will run this daemon every 5-10 minutes. It will scan for
// `*.dtf` files and compress them as `*.dtf.xz` files. From here, we will then upload them to AWS s3

// TODO: Perhaps send the files via some sort of socket server and compress with paq8px? Think about this.

// CompressDaemon : Scheduled dtf compressor
func CompressDaemon() {

}
