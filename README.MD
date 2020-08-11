# log_read_and_write

用于日志文件读写，打开两个句柄，读写分离，通过信号通知读取消费，各自维护offset，提升读写性能
