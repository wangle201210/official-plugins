# 编译静态库说明

## 编译静态库

使用 Makefile 编译静态库：

```bash
cd watermark
make clean  # 清理之前的编译文件
make        # 编译静态库 libwatermark.a
```

## 使用静态库

Go 代码会自动链接 `libwatermark.a` 静态库。确保：

1. 静态库文件 `libwatermark.a` 在 `watermark/` 目录下
2. 头文件 `watermark.h` 在 `watermark/` 目录下
3. CGO 配置中已包含 `-lwatermark` 链接选项

## 清理

```bash
make clean  # 删除编译生成的文件
```

