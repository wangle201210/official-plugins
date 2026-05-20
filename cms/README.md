# LinaPro CMS Plugin

The CMS source plugin provides content site capabilities for LinaPro. It owns
site settings, category trees, articles, links, slides, visitor messages, and
public HTML rendering APIs.

## Scope

- Source plugin under `apps/lina-plugins/cms`
- Plugin-owned `plugin_cms_*` tables
- Authenticated management APIs under `/api/v1/cms/*`
- Public read-only APIs under `/api/v1/cms/public/*`
- Public HTML pages under `/cms-site`
- Vben management page embedded through the plugin runtime
- Plugin-owned E2E tests under `hack/tests`

## Root Domain OpenResty Proxy

The public CMS site is served by LinaPro under `/cms-site`. To publish it at a
dedicated domain root, such as `https://cms.example.com/`, configure OpenResty
to proxy the domain root to the backend `/cms-site` path and rewrite generated
`/cms-site` links back to `/`.

A 1Panel/OpenResty reference server block is provided at
`deploy/openresty/cms-site-root-domain.conf`. Copy it into the 1Panel OpenResty
`conf.d` directory, replace `cms.example.com` with the real domain, and update
`127.0.0.1:9120` if LinaPro is exposed on a different host or port. Validate and
reload OpenResty after changing the file:

```bash
openresty -t
openresty -s reload
```

The reference config keeps `/cms-site` working as a redirect to `/`, proxies
`/assets/*` and page routes back to `/cms-site/*`, and uses `sub_filter` so
templates that render `/cms-site/...` links work correctly at the root domain.

## Public Templates

Public templates live in `public/templates`. They use CMS-owned template tags
with the `cms:` prefix. Static assets used by the reference site live in
`public/assets`.

### Template Files

| File | Purpose |
| --- | --- |
| `partials.html` | Shared `head`, navigation, sidebar, and footer fragments |
| `index.html` | Public home page |
| `list.html` | Standard article list page |
| `list-card.html` | Card-style article list page |
| `search.html` | Public search result page |
| `detail.html` | Article detail page |
| `single.html` | Single-page category |
| `message.html` | Visitor message page |

### Basic Syntax

| Syntax | Description |
| --- | --- |
| `{include file=comm/head.html}` | Includes a shared fragment |
| `{cms:if(condition)}...{/cms:if}` | Conditional block with `{else}` |
| `{cms:list limit=10 order=date}...{/cms:list}` | Iterates articles |
| `{cms:search limit=10 order=date}...{/cms:search}` | Iterates search results |
| `{cms:nav limit=15 parent=0}...{/cms:nav}` | Iterates categories |
| `{cms:slide limit=5 group=1}...{/cms:slide}` | Iterates slides |
| `{cms:link limit=10 group=1}...{/cms:link}` | Iterates friendly links |
| `{cms:category code=25}...{/cms:category}` | Loads one category by code |

Supported include names are `head.html`, `foot.html`, `sidebar.html`, and
`categorynav.html`.

Loop `limit` is capped at `100`. `cms:list` defaults to `12` items per page when
`limit` is omitted.

### Global Site Tags

| Tag | Output |
| --- | --- |
| `{site:path}` | Public site root path, fixed to `/cms-site` |
| `{site:assets}` | Public asset path, fixed to `/cms-site/assets` |
| `{search:action}` | Search form action, fixed to `/cms-site/search` |
| `{message:page}` | Message page URL |
| `{message:action}` | Message submit URL |
| `{site:title}`, `{site:name}` | Site name |
| `{site:subtitle}`, `{site:slogan}` | Site slogan |
| `{site:logo}` | Site logo URL |
| `{site:keywords}` | Site keywords |
| `{site:description}` | Site description |
| `{site:address}` | Contact address |
| `{site:phone}` | Contact phone |
| `{site:email}` | Contact email |
| `{site:contact}` | Contact person |
| `{site:icp}` | ICP record text |
| `{site:wechat}` | WeChat QR code URL |
| `{search:keyword}` | Current search keyword |
| `{site:year}` | Current year |
| `{category:firstlink}` | First public category link |
| `{slide:firstimage}` | First slide image |
| `{slide:firsttitle}` | First slide title |
| `{page:title}` | Current page title |
| `{page:keywords}` | Current page keywords |
| `{page:description}` | Current page description |

### Category Tags

Use `{category:*}` tags for the current category:

| Tag | Output |
| --- | --- |
| `{category:name}` | Current category name |
| `{category:link}` | Current category link |
| `{category:code}` | Current category code |
| `{category:topcode}` | Current top-level category code |
| `{category:description}` | Current category description |

Use bracket tags inside `cms:nav`, `cms:children`, `cms:grandchildren`, and `cms:category`
loops:

| Tag | Output |
| --- | --- |
| `[*:link]` | Category link |
| `[*:code]` | Category code |
| `[*:name]` | Category name |
| `[nav:title]`, `[category:title]` | Category title |
| `[nav:keywords]`, `[category:keywords]` | Category keywords |
| `[nav:description]`, `[category:description]` | Category description |
| `[nav:childcount]`, `[category:childcount]` | Child category count |
| `[nav:active]`, `[category:active]` | Active category flag |

`*` represents the active loop prefix, such as `nav`, `child`, `grandchild`, or
`category`.

Common navigation example:

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

### Article List Tags

`cms:list` supports these attributes:

| Attribute | Description |
| --- | --- |
| `code` | Category code. When omitted, the current category is used |
| `limit` | Item count, capped at `100` |
| `order` | Sort order: `id`, `date`, `manual`, or `views` |

Tags available inside `cms:list`:

| Tag | Output |
| --- | --- |
| `[list:id]` | Article ID |
| `[list:index]` | One-based item index |
| `[list:link]` | Article detail link |
| `[list:title]` | Article title |
| `[list:subtitle]` | Article subtitle |
| `[list:summary]` | Article summary |
| `[list:content]` | Article summary |
| `[list:image]` | Cover image URL |
| `[list:date]` | Publish time |
| `[list:views]` | View count |
| `[list:category]` | Category name |

Text tags such as title and description support `length` and `more` modifiers:

```html
{cms:list code=19 limit=6 order=date}
<a href="[list:link]">[list:title length=34]</a>
<span>[list:date style=Y-m-d]</span>
{/cms:list}
```

### Search Result Tags

The public search page is `/cms-site/search`, and its template is
`public/templates/search.html`. Search forms should submit with `GET` to
`{search:action}` and use `keyword` as the query parameter:

```html
<form action="{search:action}" method="get">
  <input type="text" name="keyword" value="{search:keyword}">
  <button type="submit">Search</button>
</form>
```

Search matches all published articles by title, subtitle, summary, body, tags,
SEO keywords, and SEO description. The search page is not limited to the current
category. Use the loop `code` attribute when a template needs a category
restriction.

`cms:search` supports these attributes:

| Attribute | Description |
| --- | --- |
| `code` | Category code. When omitted, all categories are searched |
| `limit` | Items per page, capped at `100` |
| `order` | Sort order: `id`, `date`, `manual`, or `views` |

Tags available inside `cms:search` match `cms:list` with the `search` prefix:

| Tag | Output |
| --- | --- |
| `[search:id]` | Article ID |
| `[search:index]` | One-based item index |
| `[search:link]` | Article detail link |
| `[search:title]` | Article title |
| `[search:subtitle]` | Article subtitle |
| `[search:summary]` | Article summary |
| `[search:content]` | Article summary |
| `[search:preview]` | Safe excerpt around the matched keyword |
| `[search:image]` | Cover image URL |
| `[search:date]` | Publish time |
| `[search:views]` | View count |
| `[search:category]` | Category name |

```html
{cms:search limit=10 order=date}
<a href="[search:link]">[search:title length=48]</a>
<p>[search:preview]</p>
<span>[search:date style=Y-m-d]</span>
{/cms:search}
```

### Article Detail Tags

Detail pages support these `{article:*}` tags:

| Tag | Output |
| --- | --- |
| `{article:title}` | Article title |
| `{article:subtitle}` | Article subtitle |
| `{article:summary}` | Article summary |
| `{article:image}` | Cover image URL |
| `{article:author}` | Author |
| `{article:source}` | Source |
| `{article:date}` | Publish time |
| `{article:views}` | View count |
| `{article:content}` | Body HTML |
| `{article:previous}` | Previous article link |
| `{article:next}` | Next article link |

`{article:date style=Y-m-d}` can be used where a formatted publish date is
needed. It currently renders the same value as `{article:date}`.

### Slide and Link Tags

Carousel slides and friendly links are maintained in the CMS management page
through the "Slides" and "Links" tabs. The `group` attribute maps to the
management form's group code. For example, the homepage carousel can use
`group=1`, while footer links can be rendered as several `group=1`, `group=2`, and
similar groups.

`cms:slide` supports these attributes:

| Attribute | Description |
| --- | --- |
| `limit` | Item count, capped at `100` |
| `group` | Group code. When omitted, all enabled slides are used |

Tags available inside `cms:slide`:

| Tag | Output |
| --- | --- |
| `[slide:index]` | One-based item index |
| `[slide:url]` | Target link |
| `[slide:image]` | Slide image URL |
| `[slide:title]` | Slide title |
| `[slide:subtitle]` | Slide subtitle |

`cms:link` supports these attributes:

| Attribute | Description |
| --- | --- |
| `limit` | Item count, capped at `100` |
| `group` | Group code. When omitted, all enabled friendly links are used |

Tags available inside `cms:link`:

| Tag | Output |
| --- | --- |
| `[link:url]` | Link URL |
| `[link:name]` | Link name |
| `[link:logo]` | Link logo URL |

Example:

```html
{cms:slide limit=5 group=1}
<a href="[slide:url]"><img src="[slide:image]" alt="[slide:title]"></a>
{/cms:slide}

{cms:link limit=10 group=1}
<a href="[link:url]" target="_blank">[link:name]</a>
{/cms:link}
```

### Pagination, Breadcrumb, and Message State

| Tag | Output |
| --- | --- |
| `{page:breadcrumb separator='&gt;'}` | Breadcrumb navigation |
| `{page:total}` | Total rows in the current list |
| `{page:first}` | First page link |
| `{page:previous}` | Previous page link |
| `{page:next}` | Next page link |
| `{page:last}` | Last page link |
| `{page:numbers}` | Numeric pagination HTML |
| `{message:submitted}` | Message submitted successfully |
| `{message:invalid}` | Message payload is invalid |
| `{message:error}` | Message submit failed |

### Conditional Tags

Condition tags only support expressions recognized by the template compiler.
They do not execute arbitrary scripts. Common forms:

```html
{cms:if(0=='{category:code}')}active{/cms:if}
{cms:if([list:image])}<img src="[list:image]" alt="[list:title]">{/cms:if}
{cms:if([nav:childcount]>0)}...{/cms:if}
{cms:if({page:total}>0)}...{else}No content{/cms:if}
{cms:if({message:submitted})}<p>Submitted</p>{/cms:if}
```

## Development

The plugin is generated and tested as an independent source plugin. Database
code generation uses `backend/hack/config.yaml`, and the generated DAO, DO, and
Entity files stay inside the plugin `backend/internal` tree.
