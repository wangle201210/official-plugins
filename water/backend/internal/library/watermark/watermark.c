#include "watermark.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include <libavcodec/avcodec.h>
#include <libavformat/avformat.h>
#include <libavutil/avutil.h>
#include <libavutil/imgutils.h>
#include <libavutil/opt.h>
#include <libavutil/log.h>
#include <libavutil/error.h>
#include <libavutil/pixdesc.h>
#include <libswscale/swscale.h>
#include <libavfilter/avfilter.h>
#include <libavfilter/buffersrc.h>
#include <libavfilter/buffersink.h>

#define MAX_OUTPUT_SIZE (50 * 1024 * 1024)  // 50MB 最大输出

static int encode_jpeg(AVFrame *frame, unsigned char **output, int *output_size, int quality) {
    AVCodec *codec = avcodec_find_encoder(AV_CODEC_ID_MJPEG);
    if (!codec) {
        return AVERROR(ENOENT);
    }

    AVCodecContext *ctx = avcodec_alloc_context3(codec);
    if (!ctx) {
        return AVERROR(ENOMEM);
    }

    ctx->pix_fmt = AV_PIX_FMT_YUVJ420P;
    ctx->width = frame->width;
    ctx->height = frame->height;
    ctx->time_base = (AVRational){1, 25};
    ctx->framerate = (AVRational){25, 1};
    
    // 设置高质量 JPEG
    ctx->flags |= AV_CODEC_FLAG_QSCALE;
    ctx->global_quality = FF_QP2LAMBDA * quality;  // quality: 1-31, 1 最高质量
    
    // 使用高质量编码参数
    av_opt_set(ctx->priv_data, "q", "2", 0);  // 最高质量
    av_opt_set(ctx->priv_data, "huffman", "1", 0);  // 启用霍夫曼编码

    int ret = avcodec_open2(ctx, codec, NULL);
    if (ret < 0) {
        avcodec_free_context(&ctx);
        return ret;
    }

    AVPacket *pkt = av_packet_alloc();
    if (!pkt) {
        avcodec_free_context(&ctx);
        return AVERROR(ENOMEM);
    }

    ret = avcodec_send_frame(ctx, frame);
    if (ret < 0) {
        av_packet_free(&pkt);
        avcodec_free_context(&ctx);
        return ret;
    }

    ret = avcodec_receive_packet(ctx, pkt);
    if (ret < 0) {
        av_packet_free(&pkt);
        avcodec_free_context(&ctx);
        return ret;
    }

    // 分配输出缓冲区
    *output = (unsigned char *)malloc(pkt->size);
    if (!*output) {
        av_packet_free(&pkt);
        avcodec_free_context(&ctx);
        return AVERROR(ENOMEM);
    }

    memcpy(*output, pkt->data, pkt->size);
    *output_size = pkt->size;

    av_packet_free(&pkt);
    avcodec_free_context(&ctx);
    return 0;
}

int process_jpg_watermark(const unsigned char *input, 
                          int input_size,
                          const char *filter_descr,
                          unsigned char *output,
                          int *output_size) {
    int ret = 0;
    AVFilterContext *buffersrc_ctx = NULL;
    AVFilterContext *buffersink_ctx = NULL;
    AVFilterGraph *filter_graph = NULL;
    AVFrame *frame_in = NULL;
    AVFrame *frame_out = NULL;
    AVCodecContext *dec_ctx = NULL;
    AVPacket *pkt = NULL;
    struct SwsContext *sws_ctx = NULL;
    unsigned char *jpeg_output = NULL;
    int jpeg_size = 0;

    // 初始化 FFmpeg
    av_log_set_level(AV_LOG_WARNING);  // 改为 WARNING 以便看到错误信息

    AVCodec *decoder = avcodec_find_decoder(AV_CODEC_ID_MJPEG);
    if (!decoder) {
        ret = AVERROR(ENOENT);
        goto end;
    }

    dec_ctx = avcodec_alloc_context3(decoder);
    if (!dec_ctx) {
        ret = AVERROR(ENOMEM);
        goto end;
    }

    ret = avcodec_open2(dec_ctx, decoder, NULL);
    if (ret < 0) {
        goto end;
    }

    pkt = av_packet_alloc();
    if (!pkt) {
        ret = AVERROR(ENOMEM);
        goto end;
    }

    frame_in = av_frame_alloc();
    if (!frame_in) {
        ret = AVERROR(ENOMEM);
        goto end;
    }

    // 解码 JPEG（直接使用内存数据）
    // 注意：需要复制数据，因为 pkt 可能被修改
    pkt->data = (unsigned char *)av_malloc(input_size + AV_INPUT_BUFFER_PADDING_SIZE);
    if (!pkt->data) {
        ret = AVERROR(ENOMEM);
        goto end;
    }
    memcpy(pkt->data, input, input_size);
    memset(pkt->data + input_size, 0, AV_INPUT_BUFFER_PADDING_SIZE);
    pkt->size = input_size;

    ret = avcodec_send_packet(dec_ctx, pkt);
    if (ret < 0) {
        goto end;
    }

    ret = avcodec_receive_frame(dec_ctx, frame_in);
    if (ret < 0) {
        goto end;
    }

    int width = frame_in->width;
    int height = frame_in->height;
    enum AVPixelFormat pix_fmt = frame_in->format;

    // 创建 filter graph
    filter_graph = avfilter_graph_alloc();
    if (!filter_graph) {
        ret = AVERROR(ENOMEM);
        goto end;
    }

    // 创建 buffersrc filter
    char args[512];
    snprintf(args, sizeof(args),
             "video_size=%dx%d:pix_fmt=%d:time_base=1/25:pixel_aspect=1/1",
             width, height, pix_fmt);

    const AVFilter *buffersrc = avfilter_get_by_name("buffer");
    ret = avfilter_graph_create_filter(&buffersrc_ctx, buffersrc, "in",
                                       args, NULL, filter_graph);
    if (ret < 0) {
        goto end;
    }

    // 创建 buffersink filter
    const AVFilter *buffersink = avfilter_get_by_name("buffersink");
    ret = avfilter_graph_create_filter(&buffersink_ctx, buffersink, "out",
                                       NULL, NULL, filter_graph);
    if (ret < 0) {
        goto end;
    }

    // 不限制输出格式，让 filter graph 自动处理
    // 如果需要特定格式，可以在最后编码时转换

    // 解析 filter 描述
    AVFilterInOut *outputs = avfilter_inout_alloc();
    AVFilterInOut *inputs = avfilter_inout_alloc();
    if (!outputs || !inputs) {
        ret = AVERROR(ENOMEM);
        goto end;
    }

    outputs->name = av_strdup("in");
    outputs->filter_ctx = buffersrc_ctx;
    outputs->pad_idx = 0;
    outputs->next = NULL;

    inputs->name = av_strdup("out");
    inputs->filter_ctx = buffersink_ctx;
    inputs->pad_idx = 0;
    inputs->next = NULL;

    ret = avfilter_graph_parse_ptr(filter_graph, filter_descr,
                                   &inputs, &outputs, NULL);
    avfilter_inout_free(&inputs);
    avfilter_inout_free(&outputs);
    if (ret < 0) {
        goto end;
    }

    ret = avfilter_graph_config(filter_graph, NULL);
    if (ret < 0) {
        goto end;
    }

    // 注意：不要在这里转换格式，让 filter graph 处理格式转换
    // 需要复制帧数据，因为原始帧会被 filter 使用
    AVFrame *frame_copy = av_frame_alloc();
    if (!frame_copy) {
        ret = AVERROR(ENOMEM);
        goto end;
    }

    frame_copy->format = frame_in->format;
    frame_copy->width = frame_in->width;
    frame_copy->height = frame_in->height;
    ret = av_frame_get_buffer(frame_copy, 32);
    if (ret < 0) {
        av_frame_free(&frame_copy);
        goto end;
    }

    av_frame_copy(frame_copy, frame_in);
    av_frame_copy_props(frame_copy, frame_in);
    
    // 设置时间戳
    frame_copy->pts = 0;
    frame_copy->pkt_dts = 0;
    frame_copy->pkt_duration = 0;

    // 推送帧到 filter
    // 让 buffersrc 复制帧（flags=0），这样我们可以在本地释放 frame_copy
    ret = av_buffersrc_add_frame_flags(buffersrc_ctx, frame_copy, 0);
    if (ret < 0) {
        av_frame_free(&frame_copy);
        goto end;
    }
 
    // 发送 EOF 信号，确保所有 filter 处理完成
    ret = av_buffersrc_add_frame_flags(buffersrc_ctx, NULL, 0);
    if (ret < 0 && ret != AVERROR_EOF) {
        av_frame_free(&frame_copy);
        goto end;
    }
 
    // 获取处理后的帧
    frame_out = av_frame_alloc();
    if (!frame_out) {
        av_frame_free(&frame_copy);
        ret = AVERROR(ENOMEM);
        goto end;
    }

    // 循环获取帧，直到获取到有效帧或遇到错误
    ret = av_buffersink_get_frame(buffersink_ctx, frame_out);
    while (ret == AVERROR(EAGAIN) || ret == AVERROR_EOF) {
        if (ret == AVERROR_EOF) {
            // 已经结束，尝试再次获取
            break;
        }
        // EAGAIN 表示需要更多输入，但我们已经发送了 EOF，所以应该能获取到帧
        ret = av_buffersink_get_frame(buffersink_ctx, frame_out);
    }
    
    if (ret < 0) {
        av_frame_free(&frame_copy);
        goto end;
    }
    
    av_frame_free(&frame_copy);
    
    // 如果输出格式不是 YUVJ420P 或 YUV420P，需要转换
    AVFrame *frame_for_encode = frame_out;
    if (frame_out->format != AV_PIX_FMT_YUV420P && frame_out->format != AV_PIX_FMT_YUVJ420P) {
        struct SwsContext *encode_sws = sws_getContext(
            frame_out->width, frame_out->height, frame_out->format,
            frame_out->width, frame_out->height, AV_PIX_FMT_YUV420P,
            SWS_BILINEAR, NULL, NULL, NULL);
        if (!encode_sws) {
            ret = AVERROR(ENOMEM);
            goto end;
        }

        AVFrame *frame_yuv = av_frame_alloc();
        if (!frame_yuv) {
            sws_freeContext(encode_sws);
            ret = AVERROR(ENOMEM);
            goto end;
        }

        frame_yuv->format = AV_PIX_FMT_YUV420P;
        frame_yuv->width = frame_out->width;
        frame_yuv->height = frame_out->height;
        ret = av_frame_get_buffer(frame_yuv, 32);
        if (ret < 0) {
            av_frame_free(&frame_yuv);
            sws_freeContext(encode_sws);
            goto end;
        }

        sws_scale(encode_sws, (const uint8_t *const *)frame_out->data,
                  frame_out->linesize, 0, frame_out->height,
                  frame_yuv->data, frame_yuv->linesize);

        sws_freeContext(encode_sws);
        frame_for_encode = frame_yuv;
    }

    // 编码为 JPEG（高质量）
    ret = encode_jpeg(frame_for_encode, &jpeg_output, &jpeg_size, 1);  // 1 = 最高质量
    if (ret < 0) {
        if (frame_for_encode != frame_out) {
            av_frame_free(&frame_for_encode);
        }
        goto end;
    }
    
    if (frame_for_encode != frame_out) {
        av_frame_free(&frame_for_encode);
    }

    // 检查输出缓冲区大小
    if (jpeg_size > *output_size) {
        ret = AVERROR(ENOSPC);
        goto end;
    }

    // 复制输出数据
    memcpy(output, jpeg_output, jpeg_size);
    *output_size = jpeg_size;

end:
    if (jpeg_output) {
        free(jpeg_output);
    }
    if (frame_out) {
        av_frame_free(&frame_out);
    }
    if (frame_in) {
        av_frame_free(&frame_in);
    }
    // frame_in 已经在处理过程中被释放或替换，不需要在这里释放
    if (sws_ctx) {
        sws_freeContext(sws_ctx);
    }
    if (filter_graph) {
        avfilter_graph_free(&filter_graph);
    }
    if (pkt) {
        if (pkt->data) {
            av_freep(&pkt->data);
        }
        av_packet_free(&pkt);
    }
    if (dec_ctx) {
        avcodec_free_context(&dec_ctx);
    }

    return ret;
}