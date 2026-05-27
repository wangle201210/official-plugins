// This file compiles CMS public frontend templates and translates CMS tag syntax.

package cms

import (
	"html/template"
	"io/fs"
	cmsplugin "lina-plugin-cms"
	cmssvc "lina-plugin-cms/backend/internal/service/cms"
	"path"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// publicFrontendTemplate returns the cached compiled public-template set.
func publicFrontendTemplate() (*template.Template, error) {
	publicFrontendTemplateCache.once.Do(func() {
		matches, err := fs.Glob(cmsplugin.EmbeddedFiles, publicFrontendTemplateGlob)
		if err != nil {
			publicFrontendTemplateCache.err = err
			return
		}
		tpl := template.New("")
		tpl = tpl.Funcs(template.FuncMap{"cmsCategoryArticles": publicFrontendCategoryArticles, "cmsCategoryByCode": publicFrontendCategoryByCode, "cmsCategoryChildren": publicFrontendCategoryChildren, "cmsGroupedLinks": publicFrontendGroupedLinks, "cmsGroupedSlides": publicFrontendGroupedSlides, "cmsLimit": publicFrontendLimit, "cmsOrderArticles": publicFrontendOrderArticles, "cmsRootCategory": publicFrontendRootCategory, "cmsTextLength": publicFrontendTextLength, "safeHTML": func(value template.HTML) template.HTML {
			return value
		}})
		for _, match := range matches {
			content, err := cmsplugin.EmbeddedFiles.ReadFile(match)
			if err != nil {
				publicFrontendTemplateCache.err = err
				return
			}
			compiled := compilePublicFrontendTemplate(string(content), publicFrontendRootScope)
			if _, err = tpl.New(path.Base(match)).Parse(compiled); err != nil {
				publicFrontendTemplateCache.err = err
				return
			}
		}
		publicFrontendTemplateCache.tpl = tpl
	})
	return publicFrontendTemplateCache.tpl, publicFrontendTemplateCache.err
}

// compilePublicFrontendTemplate compiles embedded CMS templates after tag translation.
func compilePublicFrontendTemplate(content string, scope publicFrontendTemplateScope) string {
	compiled := replacePublicFrontendIncludes(content)
	compiled = replacePublicFrontendLoops(compiled)
	compiled = replacePublicFrontendIfs(compiled, scope)
	compiled = replacePublicFrontendRootTags(compiled)
	compiled = replacePublicFrontendScopedTags(compiled, scope)
	return compiled
}

// replacePublicFrontendLoops translates CMS loop tags into Go template ranges.
func replacePublicFrontendLoops(content string) string {
	for {
		loc := publicFrontendLoopStartPattern.FindStringSubmatchIndex(content)
		if loc == nil {
			return content
		}
		name := content[loc[2]:loc[3]]
		attrs := publicFrontendParseLoopAttrs(content[loc[4]:loc[5]])
		closeTag := "{/cms:" + name + "}"
		closeIndex := strings.Index(content[loc[1]:], closeTag)
		if closeIndex < 0 {
			return content
		}
		bodyStart := loc[1]
		bodyEnd := loc[1] + closeIndex
		loopScope := publicFrontendLoopScope(name)
		body := compilePublicFrontendTemplate(content[bodyStart:bodyEnd], loopScope)
		compiled := publicFrontendLoopTemplate(loopScope, attrs, body)
		content = content[:loc[0]] + compiled + content[bodyEnd+len(closeTag):]
	}
}

// replacePublicFrontendIfs translates CMS conditional tags into Go template blocks.
func replacePublicFrontendIfs(content string, scope publicFrontendTemplateScope) string {
	for {
		matches := publicFrontendIfStartPattern.FindAllStringSubmatchIndex(content, -1)
		if len(matches) == 0 {
			return content
		}
		loc := matches[len(matches)-1]
		expression := content[loc[2]:loc[3]]
		closeTag := "{/cms:if}"
		closeIndex := strings.Index(content[loc[1]:], closeTag)
		if closeIndex < 0 {
			return content
		}
		bodyStart := loc[1]
		bodyEnd := loc[1] + closeIndex
		condition := publicFrontendIfCondition(expression, scope)
		ifTrue, ifFalse, hasElse := strings.Cut(content[bodyStart:bodyEnd], "{else}")
		body := compilePublicFrontendTemplate(ifTrue, scope)
		if hasElse {
			body = body + "{{else}}" + compilePublicFrontendTemplate(ifFalse, scope)
		}
		content = content[:loc[0]] + "{{if " + condition + "}}" + body + "{{end}}" + content[bodyEnd+len(closeTag):]
	}
}

// replacePublicFrontendIncludes inlines embedded template include directives.
func replacePublicFrontendIncludes(content string) string {
	return publicFrontendIncludePattern.ReplaceAllStringFunc(content, func(match string) string {
		parts := publicFrontendIncludePattern.FindStringSubmatch(match)
		if len(parts) != 2 {
			return match
		}
		fileName := strings.Trim(parts[1], " \"'")
		switch path.Base(fileName) {
		case "head.html":
			return `{{template "head" .}}`
		case "foot.html":
			return `{{template "foot" .}}`
		case "sidebar.html", "categorynav.html":
			return `{{template "sidebar" .}}`
		default:
			name := strings.TrimSuffix(path.Base(fileName), path.Ext(fileName))
			if name == "" {
				return match
			}
			return `{{template "` + name + `" .}}`
		}
	})
}

// publicFrontendLoopScope maps a CMS loop tag name to its template scope.
func publicFrontendLoopScope(name string) publicFrontendTemplateScope {
	switch name {
	case "nav":
		return publicFrontendNavScope
	case "children":
		return publicFrontendChildNavScope
	case "grandchildren":
		return publicFrontendGrandchildScope
	case "list":
		return publicFrontendListScope
	case "search":
		return publicFrontendSearchScope
	case "slide":
		return publicFrontendSlideScope
	case "link":
		return publicFrontendLinkScope
	case "category":
		return publicFrontendCategoryScope
	case "message":
		return publicFrontendMessageScope
	default:
		return publicFrontendRootScope
	}
}

// publicFrontendParseLoopAttrs parses attributes from a CMS loop tag.
func publicFrontendParseLoopAttrs(raw string) publicFrontendLoopAttrs {
	attrs := publicFrontendLoopAttrs{}
	for _, field := range strings.Fields(strings.TrimSpace(raw)) {
		key, value, ok := strings.Cut(field, "=")
		if !ok {
			continue
		}
		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		switch key {
		case "limit":
			limit, err := strconv.Atoi(value)
			if err == nil && limit > 0 {
				attrs.Limit = publicFrontendClampLoopLimit(limit)
			}
		case "code":
			attrs.Code = strings.TrimSpace(value)
		case "parent":
			attrs.Parent = strings.TrimSpace(value)
			if parentID, err := strconv.ParseInt(attrs.Parent, 10, 64); err == nil && parentID > 0 {
				attrs.ParentID = parentID
			}
		case "group":
			attrs.Group = strings.TrimSpace(value)
		case "order":
			attrs.Order = cmssvc.NormalizePublicArticleOrder(value)
		default:
			continue
		}
	}
	return attrs
}

// publicFrontendParseTextAttrs parses text clipping attributes from a template tag.
func publicFrontendParseTextAttrs(raw string) publicFrontendTextAttrs {
	attrs := publicFrontendTextAttrs{More: publicFrontendTextMore}
	for _, field := range strings.Fields(strings.TrimSpace(raw)) {
		key, value, ok := strings.Cut(field, "=")
		if !ok {
			continue
		}
		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		switch key {
		case "length":
			length, err := strconv.Atoi(value)
			if err == nil && length > 0 {
				attrs.Length = length
			}
		case "more":
			attrs.More = value
		default:
			continue
		}
	}
	return attrs
}

// publicFrontendClampLoopLimit bounds template loop limits to a safe maximum.
func publicFrontendClampLoopLimit(value int) int {
	if value < 1 {
		return 0
	}
	if value > publicFrontendMaxLoopSize {
		return publicFrontendMaxLoopSize
	}
	return value
}

// publicFrontendLoopLimit resolves the effective loop limit for a template scope.
func publicFrontendLoopLimit(scope publicFrontendTemplateScope, attrs publicFrontendLoopAttrs) int {
	if attrs.Limit > 0 {
		return attrs.Limit
	}
	switch scope {
	case publicFrontendListScope, publicFrontendSearchScope:
		return publicFrontendPageSize
	default:
		return 0
	}
}

// publicFrontendTemplateArticleAttrs finds article loop attributes declared by a template.
func publicFrontendTemplateArticleAttrs(templateFileName string, scope publicFrontendTemplateScope) publicFrontendLoopAttrs {
	content, err := cmsplugin.EmbeddedFiles.ReadFile(path.Join("public/templates", templateFileName))
	if err != nil {
		return publicFrontendLoopAttrs{Limit: publicFrontendPageSize}
	}
	matches := publicFrontendLoopStartPattern.FindAllStringSubmatch(string(content), -1)
	for _, match := range matches {
		if len(match) < 3 || publicFrontendLoopScope(match[1]) != scope {
			continue
		}
		return publicFrontendParseLoopAttrs(match[2])
	}
	return publicFrontendLoopAttrs{Limit: publicFrontendPageSize}
}

// publicFrontendTemplateListAttrs finds list loop attributes declared by a template.
func publicFrontendTemplateListAttrs(templateFileName string) publicFrontendLoopAttrs {
	return publicFrontendTemplateArticleAttrs(templateFileName, publicFrontendListScope)
}

// publicFrontendLoopTemplate returns the Go template fragment for one CMS loop scope.
func publicFrontendLoopTemplate(scope publicFrontendTemplateScope, attrs publicFrontendLoopAttrs, body string) string {
	limit := strconv.Itoa(publicFrontendLoopLimit(scope, attrs))
	switch scope {
	case publicFrontendNavScope:
		if attrs.Parent != "" && attrs.Parent != "0" {
			if attrs.Parent == "{category:topcode}" {
				return "{{range cmsCategoryChildren $ (cmsRootCategory $.CurrentCategory) " + limit + "}}" + body + "{{end}}"
			}
			return "{{range cmsCategoryChildren $ " + publicFrontendCategoryArg(attrs.Parent) + " " + limit + "}}" + body + "{{end}}"
		}
		return "{{range cmsLimit .NavCategories " + limit + "}}" + body + "{{end}}"
	case publicFrontendChildNavScope, publicFrontendGrandchildScope:
		return "{{range cmsLimit .Children " + limit + "}}" + body + "{{end}}"
	case publicFrontendListScope, publicFrontendSearchScope:
		if attrs.Code != "" {
			return "{{range cmsCategoryArticles $ " + publicFrontendStringLiteral(attrs.Code) + " " + limit + " " + publicFrontendStringLiteral(string(attrs.Order)) + "}}" + body + "{{end}}"
		}
		return "{{range cmsLimit (cmsOrderArticles .Articles " + publicFrontendStringLiteral(string(attrs.Order)) + ") " + limit + "}}" + body + "{{end}}"
	case publicFrontendSlideScope:
		if attrs.Group != "" {
			return "{{range cmsGroupedSlides $.Slides " + publicFrontendStringLiteral(attrs.Group) + " " + limit + "}}" + body + "{{end}}"
		}
		return "{{range cmsLimit .Slides " + limit + "}}" + body + "{{end}}"
	case publicFrontendLinkScope:
		if attrs.Group != "" {
			return "{{range cmsGroupedLinks $.Links " + publicFrontendStringLiteral(attrs.Group) + " " + limit + "}}" + body + "{{end}}"
		}
		return "{{range cmsLimit .Links " + limit + "}}" + body + "{{end}}"
	case publicFrontendMessageScope:
		return "{{range cmsLimit .ApprovedMessages " + limit + "}}" + body + "{{end}}"
	case publicFrontendCategoryScope:
		if attrs.Code != "" {
			if attrs.Code == "{category:topcode}" {
				return "{{with cmsRootCategory $.CurrentCategory}}" + body + "{{end}}"
			}
			if attrs.Code == "{category:code}" {
				return "{{with $.CurrentCategory}}" + body + "{{end}}"
			}
			return "{{with cmsCategoryByCode $ " + publicFrontendStringLiteral(attrs.Code) + "}}" + body + "{{end}}"
		}
		return "{{with .CurrentCategory}}" + body + "{{end}}"
	default:
		return body
	}
}

// publicFrontendStringLiteral quotes text for generated Go template expressions.
func publicFrontendStringLiteral(value string) string {
	return strconv.Quote(strings.TrimSpace(value))
}

// publicFrontendCategoryArg converts a category template argument into a Go template expression.
func publicFrontendCategoryArg(value string) string {
	switch strings.TrimSpace(value) {
	case "[nav:code]", "[child:code]", "[grandchild:code]":
		return "."
	case "{category:code}":
		return "$.CurrentCategory"
	default:
		return publicFrontendStringLiteral(value)
	}
}

// publicFrontendLimit bounds a slice inside generated template helper calls.
func publicFrontendLimit(items any, limit int) any {
	if items == nil || limit <= 0 {
		return items
	}
	value := reflect.ValueOf(items)
	switch value.Kind() {
	case reflect.Array, reflect.Slice:
		if value.Len() <= limit {
			return items
		}
		return value.Slice(0, limit).Interface()
	default:
		return items
	}
}

// publicFrontendIfCondition converts CMS conditional syntax into Go template conditions.
func publicFrontendIfCondition(expression string, scope publicFrontendTemplateScope) string {
	expr := strings.TrimSpace(expression)
	switch {
	case strings.Contains(expr, "0=='{category:code}'"), strings.Contains(expr, "'0'=='{category:code}'"):
		return "not .CurrentCategory"
	case expr == "{message:submitted}":
		return ".Submitted"
	case expr == "{message:invalid}":
		return ".InvalidMessage"
	case expr == "{message:error}":
		return ".MessageError"
	case expr == "{message:show}":
		return ".ShowMessages"
	case expr == "[message:reply]":
		return ".Reply"
	case expr == "{slide:firstimage}":
		return ".PrimarySlide.Image"
	case expr == "{site:logo}":
		return ".Site.Logo"
	case strings.Contains(expr, "page:total"):
		return "and .Pagination (gt .Pagination.Rows 0)"
	case expr == "{article:image}":
		return "and .CurrentArticle .CurrentArticle.Cover"
	case strings.Contains(expr, "childcount"):
		return ".Children"
	case strings.Contains(expr, "[slide:index]"):
		return "eq .Index 1"
	case expr == "[list:image]":
		return ".Cover"
	case expr == "[search:image]":
		return ".Cover"
	case strings.Contains(expr, "[nav:code]") && strings.Contains(expr, "{category:topcode}"):
		return ".Active"
	case strings.Contains(expr, "[nav:code]") && strings.Contains(expr, "{category:code}"):
		return ".Active"
	case strings.Contains(expr, "[child:code]") && strings.Contains(expr, "{category:code}"):
		return ".Active"
	case strings.Contains(expr, "[grandchild:code]") && strings.Contains(expr, "{category:code}"):
		return ".Active"
	case strings.Contains(expr, "nav:active"), strings.Contains(expr, "child:active"), strings.Contains(expr, "grandchild:active"), strings.Contains(expr, "category:active"):
		return ".Active"
	case strings.Contains(expr, "article:image"):
		return "and .CurrentArticle .CurrentArticle.Cover"
	default:
		return "false"
	}
}

// replacePublicFrontendRootTags replaces site-level tags in compiled public templates.
func replacePublicFrontendRootTags(content string) string {
	replacements := []string{"{site:path}", "/cms-site", "{site:assets}", "/cms-site/assets", "{search:action}", "/cms-site/search", "{message:action}", "/cms-site/messages", "{message:page}", "/cms-site/message", "{site:title}", "{{.Site.Name}}", "{site:name}", "{{.Site.Name}}", "{site:subtitle}", "{{.Site.Slogan}}", "{site:slogan}", "{{.Site.Slogan}}", "{site:logo}", "{{.Site.Logo}}", "{site:keywords}", "{{.Site.Keywords}}", "{site:description}", "{{.Site.Description}}", "{site:address}", "{{.Site.Address}}", "{site:phone}", "{{.Site.Phone}}", "{site:email}", "{{.Site.Email}}", "{site:contact}", "{{.Site.Contact}}", "{site:icp}", "{{.Site.Icp}}", "{site:wechat}", "{{.CompanyWeixin}}", "{search:keyword}", "{{.Keyword}}", "{site:year}", "{{.Year}}", "{category:firstlink}", "{{.FirstCategoryHref}}", "{slide:firstimage}", "{{.PrimarySlide.Image}}", "{slide:firsttitle}", "{{.PrimarySlide.Title}}", "{page:title}", "{{if .PageTitle}}{{.PageTitle}}-{{end}}{{.Site.Name}}", "{page:keywords}", "{{if .CurrentArticle}}{{.CurrentArticle.Keywords}}{{else if .CurrentCategory}}{{.CurrentCategory.Keywords}}{{else}}{{.Site.Keywords}}{{end}}", "{page:description}", "{{if .CurrentArticle}}{{.CurrentArticle.Description}}{{else if .CurrentCategory}}{{.CurrentCategory.Description}}{{else}}{{.Site.Description}}{{end}}", "{category:name}", "{{if .CurrentCategory}}{{.CurrentCategory.Name}}{{end}}", "{category:link}", "{{if .CurrentCategory}}{{.CurrentCategory.Href}}{{end}}", "{category:code}", "{{if .CurrentCategory}}{{.CurrentCategory.Code}}{{end}}", "{category:topcode}", "{{with cmsRootCategory .CurrentCategory}}{{.Code}}{{end}}", "{category:description}", "{{if .CurrentCategory}}{{.CurrentCategory.Description}}{{end}}", "{article:title}", "{{if .CurrentArticle}}{{.CurrentArticle.Title}}{{end}}", "{article:subtitle}", "{{if .CurrentArticle}}{{.CurrentArticle.Subtitle}}{{end}}", "{article:summary}", "{{if .CurrentArticle}}{{.CurrentArticle.Summary}}{{end}}", "{article:image}", "{{if .CurrentArticle}}{{.CurrentArticle.Cover}}{{end}}", "{article:author}", "{{if .CurrentArticle}}{{.CurrentArticle.Author}}{{end}}", "{article:source}", "{{if .CurrentArticle}}{{.CurrentArticle.Source}}{{end}}", "{article:date}", "{{if .CurrentArticle}}{{.CurrentArticle.PublishedAt}}{{end}}", "{article:views}", "{{if .CurrentArticle}}{{.CurrentArticle.Views}}{{end}}", "{article:content}", "{{if .CurrentArticle}}{{.CurrentArticle.ContentHTML}}{{end}}", "{article:previous}", "{{if .PreviousArticle}}<a href=\"{{.PreviousArticle.Href}}\">{{.PreviousArticle.Title}}</a>{{else}}无{{end}}", "{article:next}", "{{if .NextArticle}}<a href=\"{{.NextArticle.Href}}\">{{.NextArticle.Title}}</a>{{else}}无{{end}}", "{page:total}", "{{if .Pagination}}{{.Pagination.Rows}}{{else}}0{{end}}", "{page:first}", "{{if .Pagination}}{{.Pagination.IndexHref}}{{end}}", "{page:previous}", "{{if .Pagination}}{{.Pagination.PreHref}}{{end}}", "{page:next}", "{{if .Pagination}}{{.Pagination.NextHref}}{{end}}", "{page:last}", "{{if .Pagination}}{{.Pagination.LastHref}}{{end}}", "{page:numbers}", "{{if .Pagination}}{{.Pagination.NumBar}}{{end}}"}
	replaced := strings.NewReplacer(replacements...).Replace(content)
	replaced = regexp.MustCompile(`\{article:date\s+style=[^}]+\}`).ReplaceAllString(replaced, "{{if .CurrentArticle}}{{.CurrentArticle.PublishedAt}}{{end}}")
	replaced = regexp.MustCompile(`\{page:breadcrumb[^}]*\}`).ReplaceAllString(replaced, `<a href="/cms-site">首页</a>{{if .CurrentCategory}}<span class="sep">&gt;</span>{{with cmsRootCategory .CurrentCategory}}<a href="{{.Href}}">{{.Name}}</a>{{end}}{{if ne .CurrentCategory.Id (cmsRootCategory .CurrentCategory).Id}}<span class="sep">&gt;</span><span>{{.CurrentCategory.Name}}</span>{{end}}{{end}}`)
	return replaced
}

// replacePublicFrontendScopedTags replaces tags for the current loop scope.
func replacePublicFrontendScopedTags(content string, scope publicFrontendTemplateScope) string {
	switch scope {
	case publicFrontendNavScope:
		return replacePublicFrontendCategoryTags(content, "nav")
	case publicFrontendChildNavScope:
		return replacePublicFrontendCategoryTags(content, "child")
	case publicFrontendGrandchildScope:
		return replacePublicFrontendCategoryTags(content, "grandchild")
	case publicFrontendListScope:
		return replacePublicFrontendArticleTags(content, "list")
	case publicFrontendSearchScope:
		return replacePublicFrontendArticleTags(content, "search")
	case publicFrontendSlideScope:
		return replacePublicFrontendSlideTags(content)
	case publicFrontendLinkScope:
		return replacePublicFrontendLinkTags(content)
	case publicFrontendCategoryScope:
		return replacePublicFrontendCategoryTags(content, "category")
	case publicFrontendMessageScope:
		return replacePublicFrontendMessageTags(content)
	default:
		return content
	}
}

// replacePublicFrontendCategoryTags replaces category field tags.
func replacePublicFrontendCategoryTags(content string, prefix string) string {
	replaced := strings.NewReplacer("["+prefix+":link]", "{{.Href}}", "["+prefix+":code]", "{{.Code}}", "["+prefix+":name]", "{{.Name}}", "["+prefix+":title]", "{{.Title}}", "["+prefix+":keywords]", "{{.Keywords}}", "["+prefix+":description]", "{{.Description}}", "["+prefix+":childcount]", "{{len .Children}}", "["+prefix+":active]", "{{.Active}}").Replace(content)
	replaced = replacePublicFrontendTextParamTags(replaced, prefix, "name", ".Name")
	replaced = replacePublicFrontendTextParamTags(replaced, prefix, "title", ".Title")
	replaced = replacePublicFrontendTextParamTags(replaced, prefix, "description", ".Description")
	return replaced
}

// replacePublicFrontendArticleTags replaces article field tags.
func replacePublicFrontendArticleTags(content string, prefix string) string {
	replaced := strings.NewReplacer("["+prefix+":id]", "{{.Id}}", "["+prefix+":index]", "{{.Index}}", "["+prefix+":link]", "{{.Href}}", "["+prefix+":title]", "{{.Title}}", "["+prefix+":subtitle]", "{{.Subtitle}}", "["+prefix+":summary]", "{{.Summary}}", "["+prefix+":content]", "{{.Summary}}", "["+prefix+":preview]", "{{.SearchPreview}}", "["+prefix+":image]", "{{.Cover}}", "["+prefix+":date]", "{{.PublishedAt}}", "["+prefix+":views]", "{{.Views}}", "["+prefix+":category]", "{{.CategoryName}}").Replace(content)
	replaced = replacePublicFrontendTextParamTags(replaced, prefix, "title", ".Title")
	replaced = replacePublicFrontendTextParamTags(replaced, prefix, "subtitle", ".Subtitle")
	replaced = replacePublicFrontendTextParamTags(replaced, prefix, "summary", ".Summary")
	replaced = replacePublicFrontendTextParamTags(replaced, prefix, "content", ".Summary")
	replaced = regexp.MustCompile(`\[`+regexp.QuoteMeta(prefix)+`:title\s+[^\]]+\]`).ReplaceAllString(replaced, "{{.Title}}")
	replaced = regexp.MustCompile(`\[`+regexp.QuoteMeta(prefix)+`:summary\s+[^\]]+\]`).ReplaceAllString(replaced, "{{.Summary}}")
	replaced = regexp.MustCompile(`\[`+regexp.QuoteMeta(prefix)+`:content\s+[^\]]+\]`).ReplaceAllString(replaced, "{{.Summary}}")
	replaced = regexp.MustCompile(`\[`+regexp.QuoteMeta(prefix)+`:preview\s+[^\]]+\]`).ReplaceAllString(replaced, "{{.SearchPreview}}")
	replaced = regexp.MustCompile(`\[`+regexp.QuoteMeta(prefix)+`:date\s+[^\]]+\]`).ReplaceAllString(replaced, "{{.PublishedAt}}")
	return replaced
}

// replacePublicFrontendSlideTags replaces slide field tags.
func replacePublicFrontendSlideTags(content string) string {
	replaced := strings.NewReplacer("[slide:index]", "{{.Index}}", "[slide:url]", "{{.Link}}", "[slide:image]", "{{.Image}}").Replace(content)
	replaced = replacePublicFrontendTextParamTags(replaced, "slide", "title", ".Title")
	replaced = replacePublicFrontendTextParamTags(replaced, "slide", "subtitle", ".Subtitle")
	replaced = strings.NewReplacer("[slide:title]", "{{.Title}}", "[slide:subtitle]", "{{.Subtitle}}").Replace(replaced)
	return replaced
}

// replacePublicFrontendLinkTags replaces friendly-link field tags.
func replacePublicFrontendLinkTags(content string) string {
	return strings.NewReplacer("[link:url]", "{{.Url}}", "[link:name]", "{{.Name}}", "[link:logo]", "{{.Logo}}").Replace(content)
}

// replacePublicFrontendMessageTags replaces public visitor-message field tags.
func replacePublicFrontendMessageTags(content string) string {
	replaced := strings.NewReplacer("[message:id]", "{{.Id}}", "[message:name]", "{{.Name}}", "[message:content]", "{{.Content}}", "[message:reply]", "{{.Reply}}", "[message:date]", "{{.CreatedAt}}", "[message:updated]", "{{.UpdatedAt}}").Replace(content)
	replaced = replacePublicFrontendTextParamTags(replaced, "message", "name", ".Name")
	replaced = replacePublicFrontendTextParamTags(replaced, "message", "content", ".Content")
	replaced = replacePublicFrontendTextParamTags(replaced, "message", "reply", ".Reply")
	return replaced
}

// replacePublicFrontendTextParamTags replaces text clipping helper tags.
func replacePublicFrontendTextParamTags(content string, prefix string, name string, expression string) string {
	pattern := regexp.MustCompile(`\[` + regexp.QuoteMeta(prefix) + `:` + regexp.QuoteMeta(name) + `\s+([^\]]+)\]`)
	return pattern.ReplaceAllStringFunc(content, func(match string) string {
		parts := pattern.FindStringSubmatch(match)
		if len(parts) != 2 {
			return match
		}
		attrs := publicFrontendParseTextAttrs(parts[1])
		if attrs.Length > 0 {
			return "{{cmsTextLength " + expression + " " + strconv.Itoa(attrs.Length) + " " + publicFrontendStringLiteral(attrs.More) + "}}"
		}
		return "{{" + expression + "}}"
	})
}
