# LinaPro CMS 插件

CMS 源码插件为 LinaPro 提供内容站点能力。它负责站点配置、栏目树、文章内容、友情链接、轮播图、访客留言以及公开 HTML 页面渲染。

## 范围

- 源码插件目录为 `apps/lina-plugins/cms`
- 插件自有 `plugin_cms_*` 数据表
- 管理端认证接口位于 `/api/v1/cms/*`
- 公开只读接口位于 `/api/v1/cms/public/*`
- 公开 HTML 页面位于 `/cms-site`
- 通过插件运行时嵌入 Vben 管理页面
- 插件自有 E2E 测试位于 `hack/tests`

正常插件安装只加载数据结构和治理资源。Starter 站点内容放在`manifest/sql/mock-data/`下，
仅用于本地校验和显式的运行时示例加载，避免新安装环境默认写入演示内容。

用户确认 starter 内容效果后，可在 CMS 管理页面使用`清空数据`操作删除 CMS
栏目、文章、标签、轮播图、友情链接和访客留言。该操作会保留一条空白默认站点记录，
避免管理页面和公开接口在清理后不可用，方便用户继续建设自己的正式站点。共享上传文件
归宿主文件模块管理，不属于 CMS 插件业务表数据，因此不会被该操作删除。

管理员也可以使用`加载示例数据`操作清空当前 CMS 业务内容，并重新加载
`manifest/sql/mock-data/002-cms-starter-content.sql`内置的 starter 示例站点。该能力适用于
演示站点被清空或验证后需要恢复到交付示例状态的场景。

## 根域名 OpenResty 代理

CMS 公开站点在 LinaPro 后端内固定挂载于`/cms-site`。如果希望使用独立域名根路径访问，例如`https://cms.example.com/`，需要让 OpenResty 将该域名根路径代理到后端`/cms-site`，并把页面中生成的`/cms-site`链接改写为`/`。

插件提供了一份 1Panel/OpenResty 参考配置：
`deploy/openresty/cms-site-root-domain.conf`。使用时可复制到 1Panel OpenResty 的`conf.d`目录，将`cms.example.com`替换为真实域名；如果 LinaPro 后端不在`127.0.0.1:9120`，也需要同步替换代理地址。修改后执行配置检查和重载：

```bash
openresty -t
openresty -s reload
```

该参考配置会将旧的`/cms-site`访问重定向到根路径，把`/assets/*`和栏目、文章等页面路径代理回后端`/cms-site/*`，并通过`sub_filter`让模板中固定生成的`/cms-site/...`链接在根域名下正常工作。

## 公开模板

公开模板位于`public/templates`。模板标签使用 CMS 自有的`cms:`前缀，参考站点使用的静态资源位于`public/assets`。

### 模板文件

| 文件 | 用途 |
| --- | --- |
| `partials.html` | 统一的`head`、导航、侧栏与页脚片段 |
| `index.html` | 公开首页 |
| `list.html` | 普通文章列表页 |
| `list-card.html` | 卡片式文章列表页 |
| `search.html` | 公开搜索结果页 |
| `detail.html` | 文章详情页 |
| `single.html` | 单页栏目 |
| `message.html` | 访客留言页 |

### 基础语法

| 语法 | 说明 |
| --- | --- |
| `{include file=comm/head.html}` | 引入公共片段 |
| `{cms:if(条件)}...{/cms:if}` | 条件渲染，支持`{else}`分支 |
| `{cms:list limit=10 order=date}...{/cms:list}` | 循环输出文章列表 |
| `{cms:search limit=10 order=date}...{/cms:search}` | 循环输出搜索结果 |
| `{cms:nav limit=15 parent=0}...{/cms:nav}` | 循环输出栏目导航 |
| `{cms:slide limit=5 group=1}...{/cms:slide}` | 循环输出轮播图 |
| `{cms:link limit=10 group=1}...{/cms:link}` | 循环输出友情链接 |
| `{cms:category code=25}...{/cms:category}` | 按栏目编号读取一个栏目 |

支持的公共片段文件名包括`head.html`、`foot.html`、`sidebar.html`和
`categorynav.html`。

循环的`limit`最大值为`100`。`cms:list`未指定`limit`时默认每页`12`条。

### 全局站点标签

| 标签 | 输出 |
| --- | --- |
| `{site:path}` | 公开站点根路径，固定为`/cms-site` |
| `{site:assets}` | 公开资源路径，固定为`/cms-site/assets` |
| `{search:action}` | 搜索表单地址，固定为`/cms-site/search` |
| `{message:page}` | 留言页面地址 |
| `{message:action}` | 留言提交地址 |
| `{site:title}`、`{site:name}` | 站点名称 |
| `{site:subtitle}`、`{site:slogan}` | 站点标语 |
| `{site:logo}` | 站点 Logo 地址 |
| `{site:keywords}` | 站点关键词 |
| `{site:description}` | 站点描述 |
| `{site:address}` | 联系地址 |
| `{site:phone}` | 联系电话 |
| `{site:email}` | 联系邮箱 |
| `{site:contact}` | 联系人 |
| `{site:icp}` | 备案信息 |
| `{site:wechat}` | 微信二维码地址 |
| `{search:keyword}` | 当前搜索关键词 |
| `{site:year}` | 当前年份 |
| `{category:firstlink}` | 第一个公开栏目的链接 |
| `{slide:firstimage}` | 第一张轮播图图片 |
| `{slide:firsttitle}` | 第一张轮播图标题 |
| `{page:title}` | 当前页面标题 |
| `{page:keywords}` | 当前页面关键词 |
| `{page:description}` | 当前页面描述 |

### 栏目标签

当前栏目可直接使用`{category:*}`标签：

| 标签 | 输出 |
| --- | --- |
| `{category:name}` | 当前栏目名称 |
| `{category:link}` | 当前栏目链接 |
| `{category:code}` | 当前栏目编号 |
| `{category:topcode}` | 当前顶级栏目编号 |
| `{category:description}` | 当前栏目描述 |

在`cms:nav`、`cms:children`、`cms:grandchildren`和`cms:category`循环内使用方括号标签：

| 标签 | 输出 |
| --- | --- |
| `[nav:link]`、`[child:link]`、`[grandchild:link]`、`[category:link]` | 栏目链接 |
| `[nav:code]`、`[child:code]`、`[grandchild:code]`、`[category:code]` | 栏目编号 |
| `[nav:name]`、`[child:name]`、`[grandchild:name]`、`[category:name]` | 栏目名称 |
| `[nav:title]`、`[category:title]` | 栏目标题 |
| `[nav:keywords]`、`[category:keywords]` | 栏目关键词 |
| `[nav:description]`、`[category:description]` | 栏目描述 |
| `[nav:childcount]`、`[category:childcount]` | 子栏目数量 |
| `[nav:active]`、`[category:active]` | 当前栏目或其父栏目是否处于激活状态 |

常用导航示例：

```html
{cms:nav limit=15 parent=0}
<li>
  <a href="[nav:link]" class="{cms:if('[nav:code]'=='{category:topcode}')}active{/cms:if}">[nav:name]</a>
  {cms:if([nav:childcount]>0)}
  <ul>
    {cms:children parent=[nav:code]}
    <li><a href="[child:link]">[child:name]</a></li>
    {/cms:children}
  </ul>
  {/cms:if}
</li>
{/cms:nav}
```

### 文章列表标签

`cms:list`支持以下属性：

| 属性 | 说明 |
| --- | --- |
| `code` | 栏目编号；不指定时使用当前栏目 |
| `limit` | 输出数量，最大`100` |
| `order` | 排序方式，支持`id`、`date`、`manual`、`views` |

`cms:list`循环内可用标签：

| 标签 | 输出 |
| --- | --- |
| `[list:id]` | 文章 ID |
| `[list:index]` | 当前序号，从`1`开始 |
| `[list:link]` | 文章详情链接 |
| `[list:title]` | 文章标题 |
| `[list:subtitle]` | 文章副标题 |
| `[list:summary]` | 文章摘要 |
| `[list:content]` | 文章摘要 |
| `[list:image]` | 封面图地址 |
| `[list:date]` | 发布时间 |
| `[list:views]` | 浏览量 |
| `[list:category]` | 所属栏目名称 |

标题、摘要等文本标签支持`length`和`more`参数：

```html
{cms:list code=19 limit=6 order=date}
<a href="[list:link]">[list:title length=34]</a>
<span>[list:date style=Y-m-d]</span>
{/cms:list}
```

### 搜索结果标签

公开搜索页位于`/cms-site/search`，模板文件为`public/templates/search.html`。
表单应使用`GET`方式提交到`{search:action}`，关键词参数名为`keyword`：

```html
<form action="{search:action}" method="get">
  <input type="text" name="keyword" value="{search:keyword}">
  <button type="submit">搜索</button>
</form>
```

搜索会在全部已发布文章中检索标题、副标题、摘要、正文、标签、SEO 关键词和
SEO 描述。搜索结果页不限定当前栏目；如需限定栏目，可在模板循环上使用
`code`参数。

`cms:search`支持以下属性：

| 属性 | 说明 |
| --- | --- |
| `code` | 栏目编号；不指定时搜索全部栏目 |
| `limit` | 每页输出数量，最大`100` |
| `order` | 排序方式，支持`id`、`date`、`manual`、`views` |

`cms:search`循环内可用标签与`cms:list`一致，只是前缀为`search`：

| 标签 | 输出 |
| --- | --- |
| `[search:id]` | 文章 ID |
| `[search:index]` | 当前序号，从`1`开始 |
| `[search:link]` | 文章详情链接 |
| `[search:title]` | 文章标题 |
| `[search:subtitle]` | 文章副标题 |
| `[search:summary]` | 文章摘要 |
| `[search:content]` | 文章摘要 |
| `[search:preview]` | 命中关键词的安全预览片段 |
| `[search:image]` | 封面图地址 |
| `[search:date]` | 发布时间 |
| `[search:views]` | 浏览量 |
| `[search:category]` | 所属栏目名称 |

```html
{cms:search limit=10 order=date}
<a href="[search:link]">[search:title length=48]</a>
<p>[search:preview]</p>
<span>[search:date style=Y-m-d]</span>
{/cms:search}
```

### 文章详情标签

详情页可用以下`{article:*}`标签：

| 标签 | 输出 |
| --- | --- |
| `{article:title}` | 文章标题 |
| `{article:subtitle}` | 文章副标题 |
| `{article:summary}` | 文章摘要 |
| `{article:image}` | 封面图地址 |
| `{article:author}` | 作者 |
| `{article:source}` | 来源 |
| `{article:date}` | 发布时间 |
| `{article:views}` | 浏览量 |
| `{article:content}` | 正文 HTML |
| `{article:previous}` | 上一篇链接 |
| `{article:next}` | 下一篇链接 |

`{article:date style=Y-m-d}`可用于指定日期输出位置，当前输出与`{article:date}`一致。

### 轮播图与友情链接标签

轮播图和友情链接数据在后台 CMS 管理页的“轮播图”和“友情链接”页签维护。
标签中的`group`对应后台表单里的“分组编码”，例如首页轮播默认使用
`group=1`，页脚友情链接可以按`group=1`、`group=2`等分组展示。

`cms:slide`支持以下属性：

| 属性 | 说明 |
| --- | --- |
| `limit` | 输出数量，最大`100` |
| `group` | 分组编码；不指定时读取全部已启用轮播图 |

`cms:slide`循环内可用标签：

| 标签 | 输出 |
| --- | --- |
| `[slide:index]` | 当前序号，从`1`开始 |
| `[slide:url]` | 跳转链接 |
| `[slide:image]` | 轮播图片地址 |
| `[slide:title]` | 轮播标题 |
| `[slide:subtitle]` | 轮播副标题 |

`cms:link`支持以下属性：

| 属性 | 说明 |
| --- | --- |
| `limit` | 输出数量，最大`100` |
| `group` | 分组编码；不指定时读取全部已启用友情链接 |

`cms:link`循环内可用标签：

| 标签 | 输出 |
| --- | --- |
| `[link:url]` | 链接地址 |
| `[link:name]` | 链接名称 |
| `[link:logo]` | 链接标识图片地址 |

示例：

```html
{cms:slide limit=5 group=1}
<a href="[slide:url]"><img src="[slide:image]" alt="[slide:title]"></a>
{/cms:slide}

{cms:link limit=10 group=1}
<a href="[link:url]" target="_blank">[link:name]</a>
{/cms:link}
```

### 分页、面包屑与留言状态

公开留言列表在默认演示数据中开启。在 CMS 站点配置中关闭`留言是否展示`后，
`/cms-site/message`会隐藏已审核通过的留言和公开回复。

| 标签 | 输出 |
| --- | --- |
| `{page:breadcrumb separator='&gt;'}` | 面包屑导航 |
| `{page:total}` | 当前列表总记录数 |
| `{page:first}` | 首页分页链接 |
| `{page:previous}` | 上一页链接 |
| `{page:next}` | 下一页链接 |
| `{page:last}` | 末页链接 |
| `{page:numbers}` | 数字分页 HTML |
| `{message:submitted}` | 留言提交成功状态 |
| `{message:invalid}` | 留言参数非法状态 |
| `{message:error}` | 留言提交失败状态 |
| `{message:show}` | 已开启审核通过留言的公开展示 |
| `{cms:message limit=12}` | 循环展示已审核通过的访客留言，最多 `50` 条 |
| `[message:id]` | 留言 ID |
| `[message:name]` | 访客名称 |
| `[message:content]` | 留言内容 |
| `[message:reply]` | 公开回复内容 |
| `[message:date]` | 留言创建时间 |
| `[message:updated]` | 留言更新时间 |

### 条件标签

条件标签只支持模板编译器内置的表达式，不支持任意脚本。常用写法如下：

```html
{cms:if(0=='{category:code}')}active{/cms:if}
{cms:if([list:image])}<img src="[list:image]" alt="[list:title]">{/cms:if}
{cms:if([nav:childcount]>0)}...{/cms:if}
{cms:if({page:total}>0)}...{else}暂无内容{/cms:if}
{cms:if({message:submitted})}<p>提交成功</p>{/cms:if}
{cms:if({message:show})}
  {cms:message limit=12}
    <article>
      <h3>[message:name]</h3>
      <p>[message:content]</p>
      {cms:if([message:reply])}<p>[message:reply]</p>{/cms:if}
    </article>
  {/cms:message}
{/cms:if}
```

## 开发

该插件按独立源码插件生成与测试。数据库代码生成使用
`backend/hack/config.yaml`，生成的 DAO、DO、Entity 文件均保留在插件自己的
`backend/internal`目录下。
