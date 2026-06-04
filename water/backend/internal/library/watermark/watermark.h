#ifndef WATERMARK_H
#define WATERMARK_H

#ifdef __cplusplus
extern "C" {
#endif

/**
 * @brief 处理 JPG 图片水印
 * @param input 输入图片数据
 * @param input_size 输入图片数据大小
 * @param filter_descr 过滤器描述
 * @param output 输出图片数据指针（需要预先分配足够大的内存）
 * @param output_size 输出图片数据大小指针（输入时表示缓冲区大小，输出时表示实际数据大小）
 * @return 0 成功，其他 失败
 */
int process_jpg_watermark(const unsigned char *input, 
                          int input_size,
                          const char *filter_descr,
                          unsigned char *output,
                          int *output_size);

#ifdef __cplusplus
}
#endif

#endif // WATERMARK_H