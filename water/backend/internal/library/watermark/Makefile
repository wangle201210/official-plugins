# Makefile for building watermark static library

CC = gcc
AR = ar
CFLAGS = -I. -I/usr/local/include $(shell pkg-config --cflags libavformat libavcodec libavutil libavfilter libswscale)
LDFLAGS = -L/usr/local/lib $(shell pkg-config --libs libavformat libavcodec libavutil libavfilter libswscale)

# 静态库名称
LIBNAME = libwatermark.a
OBJ = watermark.o

# 默认目标
all: $(LIBNAME)

# 编译静态库
$(LIBNAME): $(OBJ)
	$(AR) rcs $@ $^

# 编译目标文件
$(OBJ): watermark.c watermark.h
	$(CC) $(CFLAGS) -c watermark.c -o $@

# 清理
clean:
	rm -f $(OBJ) $(LIBNAME)

.PHONY: all clean

