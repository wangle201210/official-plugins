<script lang="ts">
export const pluginPageMeta = {
  routePath: "/cms",
  title: "CMS",
};
</script>

<script setup lang="ts">
import type { Article, Category, Link, Message, Site, Slide } from "./cms-client";

import { computed, onMounted, reactive, ref, watch } from "vue";

import { Page } from "@vben/common-ui";

import {
  Button as AButton,
  Empty as AEmpty,
  Form as AForm,
  FormItem as AFormItem,
  Input as AInput,
  InputNumber as AInputNumber,
  message,
  Modal as AModal,
  Popconfirm,
  Select as ASelect,
  Space,
  TabPane as ATabPane,
  Table as ATable,
  Tabs as ATabs,
  Tag as ATag,
  Textarea as ATextarea,
} from "ant-design-vue";

import { DictTag } from "#/components/dict";
import { $t } from "#/locales";
import { useDictStore } from "#/store/dict";

import CmsImageUpload from "../components/CmsImageUpload.vue";
import CmsRichTextEditor from "../components/CmsRichTextEditor.vue";

import {
  cmsArticleCreate,
  cmsArticleDelete,
  cmsArticleInfo,
  cmsArticleList,
  cmsArticleUpdate,
  cmsCategoryCreate,
  cmsCategoryDelete,
  cmsCategoryList,
  cmsCategoryUpdate,
  cmsLinkCreate,
  cmsLinkDelete,
  cmsLinkList,
  cmsLinkUpdate,
  cmsMessageDelete,
  cmsMessageList,
  cmsMessageUpdate,
  cmsSite,
  cmsSiteUpdate,
  cmsSlideCreate,
  cmsSlideDelete,
  cmsSlideList,
  cmsSlideUpdate,
} from "./cms-client";

type SelectNumberOption = {
  disabled?: boolean;
  label: string;
  value: number;
};

const dictStore = useDictStore();
const activeTab = ref("site");
const dashboardLoading = ref(false);

const siteForm = reactive<Partial<Site>>({
  address: "",
  contact: "",
  description: "",
  domain: "",
  email: "",
  icp: "",
  keywords: "",
  logo: "",
  name: "",
  phone: "",
  slogan: "",
  status: 1,
  weixin: "",
});

const categoryRows = ref<Category[]>([]);
const categoryModalOpen = ref(false);
const categoryModalMode = ref<"create" | "update">("create");
const categoryForm = reactive<Partial<Category>>({
  code: "",
  contentTemplate: "detail.html",
  cover: "",
  description: "",
  keywords: "",
  listTemplate: "list.html",
  name: "",
  outlink: "",
  parentId: 0,
  path: "",
  sort: 0,
  status: 1,
  title: "",
  type: 1,
});

const articleRows = ref<Article[]>([]);
const articleTotal = ref(0);
const allArticleTotal = ref(0);
const articleLoading = ref(false);
const articleQuery = reactive({
  categoryId: undefined as number | undefined,
  categoryType: undefined as number | undefined,
  includeChildren: true,
  pageNum: 1,
  pageSize: 10,
  status: undefined as number | undefined,
  title: "",
});
const selectedArticleSection = ref("model:list");
const expandedArticleCategoryIds = ref<Set<number>>(new Set());
const articleModalOpen = ref(false);
const articleModalMode = ref<"create" | "update">("create");
const articleForm = reactive<Partial<Article>>({
  author: "",
  categoryId: undefined,
  content: "",
  cover: "",
  description: "",
  isRecommend: 0,
  isTop: 0,
  keywords: "",
  slug: "",
  sort: 0,
  source: "",
  status: 0,
  subtitle: "",
  summary: "",
  tags: "",
  title: "",
});
const publishedArticleTotal = ref(0);
const draftArticleTotal = ref(0);

const messageRows = ref<Message[]>([]);
const messageTotal = ref(0);
const messageLoading = ref(false);
const messageQuery = reactive({
  keyword: "",
  pageNum: 1,
  pageSize: 10,
  status: undefined as number | undefined,
});
const pendingMessageTotal = ref(0);
const messageModalOpen = ref(false);
const messageForm = reactive<Partial<Message>>({
  reply: "",
  status: 1,
});
const currentMessage = ref<Message | null>(null);

const slideRows = ref<Slide[]>([]);
const slideTotal = ref(0);
const slideLoading = ref(false);
const slideQuery = reactive({
  groupCode: "",
  keyword: "",
  pageNum: 1,
  pageSize: 10,
  status: undefined as number | undefined,
});
const slideModalOpen = ref(false);
const slideModalMode = ref<"create" | "update">("create");
const slideForm = reactive<Partial<Slide>>({
  groupCode: "1",
  image: "",
  link: "",
  sort: 0,
  status: 1,
  subtitle: "",
  title: "",
});

const linkRows = ref<Link[]>([]);
const linkTotal = ref(0);
const linkLoading = ref(false);
const linkQuery = reactive({
  groupCode: "",
  keyword: "",
  pageNum: 1,
  pageSize: 10,
  status: undefined as number | undefined,
});
const linkModalOpen = ref(false);
const linkModalMode = ref<"create" | "update">("create");
const linkForm = reactive<Partial<Link>>({
  groupCode: "1",
  logo: "",
  name: "",
  sort: 0,
  status: 1,
  url: "",
});

const statusDicts = ref<any[]>([]);
const categoryTypeDicts = ref<any[]>([]);
const articleStatusDicts = ref<any[]>([]);
const messageStatusDicts = ref<any[]>([]);
const yesNoDicts = ref<any[]>([]);

const CategoryTypeList = 1;
const CategoryTypeSingle = 2;
const CategoryTypeExternal = 3;

const categoryOptions = computed(() =>
  flattenCategories(
    categoryRows.value,
    "",
    categoryModalMode.value === "update" ? categoryForm.id : undefined,
  ),
);
const articleCategoryOptions = computed(() =>
  flattenCategories(
    categoryRows.value.filter(
      (category) => category.type !== CategoryTypeExternal,
    ),
  ).filter((item) => item.value > 0),
);

const siteStatusOptions = computed(() => toNumberOptions(statusDicts.value));
const categoryTypeOptions = computed(() =>
  toNumberOptions(categoryTypeDicts.value),
);
const articleStatusOptions = computed(() =>
  toNumberOptions(articleStatusDicts.value),
);
const messageStatusOptions = computed(() =>
  toNumberOptions(messageStatusDicts.value),
);
const yesNoOptions = computed(() => toNumberOptions(yesNoDicts.value));
const categoryListTemplateOptions = computed(() => [
  {
    label: $t("plugin.cms.templates.list"),
    value: "list.html",
  },
  {
    label: $t("plugin.cms.templates.listCard"),
    value: "list-card.html",
  },
]);
const categoryContentTemplateOptions = computed(() => [
  {
    label: $t("plugin.cms.templates.detail"),
    value: "detail.html",
  },
  {
    label: $t("plugin.cms.templates.single"),
    value: "single.html",
  },
]);

const categoryTotal = computed(() => countCategories(categoryRows.value));
const enabledCategoryTotal = computed(() =>
  countCategories(categoryRows.value, (category) => category.status === 1),
);
const articleCategoryTree = computed(() =>
  filterArticleCategoryTree(categoryRows.value),
);
const articleCategoryNavItems = computed(() =>
  flattenVisibleArticleCategoryNavItems(
    articleCategoryTree.value,
    expandedArticleCategoryIds.value,
  ),
);
const articleSectionGroups = computed(() => [
  {
    categoryType: CategoryTypeSingle,
    key: "model:single",
    label: $t("plugin.cms.contentModels.single"),
    tone: "amber",
    total: countCategoriesByType(categoryRows.value, CategoryTypeSingle),
  },
  {
    categoryType: CategoryTypeList,
    key: "model:list",
    label: $t("plugin.cms.contentModels.list"),
    tone: "blue",
    total: countCategoriesByType(categoryRows.value, CategoryTypeList),
  },
]);
const selectedArticleSectionLabel = computed(() => {
  const model = articleSectionGroups.value.find(
    (item) => item.key === selectedArticleSection.value,
  );
  if (model) {
    return model.label;
  }
  const category = findCategoryById(categoryRows.value, articleQuery.categoryId);
  return category?.name || $t("plugin.cms.sections.contentLibrary");
});
const latestArticles = computed(() => articleRows.value.slice(0, 5));
const pendingMessages = computed(() =>
  messageRows.value.filter((item) => item.status === 0).slice(0, 4),
);
const siteDomainLabel = computed(() => siteForm.domain || "--");
const siteStatusTone = computed(() =>
  Number(siteForm.status) === 1 ? "success" : "default",
);

const metrics = computed(() => [
  {
    icon: $t("plugin.cms.metrics.categoryGlyph"),
    label: $t("plugin.cms.metrics.categories"),
    testId: "cms-metric-categories",
    tone: "teal",
    value: categoryTotal.value,
  },
  {
    icon: $t("plugin.cms.metrics.articleGlyph"),
    label: $t("plugin.cms.metrics.articles"),
    testId: "cms-metric-articles",
    tone: "blue",
    value: allArticleTotal.value,
  },
  {
    icon: $t("plugin.cms.metrics.publishedGlyph"),
    label: $t("plugin.cms.metrics.published"),
    testId: "cms-metric-published",
    tone: "green",
    value: publishedArticleTotal.value,
  },
  {
    icon: $t("plugin.cms.metrics.messageGlyph"),
    label: $t("plugin.cms.metrics.pendingMessages"),
    testId: "cms-metric-pending-messages",
    tone: "amber",
    value: pendingMessageTotal.value,
  },
]);

const categoryColumns = computed(() => [
  {
    dataIndex: "name",
    title: $t("plugin.cms.fields.categoryName"),
  },
  {
    dataIndex: "code",
    title: $t("plugin.cms.fields.categoryCode"),
    width: 180,
  },
  {
    dataIndex: "type",
    title: $t("plugin.cms.fields.categoryType"),
    width: 140,
  },
  {
    dataIndex: "status",
    title: $t("pages.common.status"),
    width: 120,
  },
  {
    dataIndex: "sort",
    title: $t("pages.fields.sort"),
    width: 96,
  },
  {
    key: "action",
    title: $t("pages.common.actions"),
    width: 180,
  },
]);

const articleColumns = computed(() => [
  {
    dataIndex: "title",
    title: $t("plugin.cms.fields.articleTitle"),
    width: 330,
  },
  {
    dataIndex: "categoryName",
    title: $t("plugin.cms.fields.categoryName"),
    width: 128,
  },
  {
    dataIndex: "status",
    title: $t("pages.common.status"),
    width: 92,
  },
  {
    dataIndex: "isRecommend",
    title: $t("plugin.cms.fields.recommend"),
    width: 70,
  },
  {
    dataIndex: "views",
    title: $t("plugin.cms.fields.views"),
    width: 62,
  },
  {
    key: "action",
    title: $t("pages.common.actions"),
    width: 136,
  },
]);

const messageColumns = computed(() => [
  {
    dataIndex: "name",
    title: $t("plugin.cms.fields.visitorName"),
    width: 140,
  },
  {
    dataIndex: "email",
    title: $t("pages.fields.email"),
    width: 210,
  },
  {
    dataIndex: "content",
    title: $t("plugin.cms.fields.messageContent"),
  },
  {
    dataIndex: "status",
    title: $t("pages.common.status"),
    width: 120,
  },
  {
    key: "action",
    title: $t("pages.common.actions"),
    width: 210,
  },
]);

const slideColumns = computed(() => [
  {
    dataIndex: "title",
    title: $t("plugin.cms.fields.slideTitle"),
    width: 280,
  },
  {
    dataIndex: "groupCode",
    title: $t("plugin.cms.fields.groupCode"),
    width: 120,
  },
  {
    dataIndex: "image",
    title: $t("plugin.cms.fields.slideImage"),
    width: 160,
  },
  {
    dataIndex: "status",
    title: $t("pages.common.status"),
    width: 120,
  },
  {
    dataIndex: "sort",
    title: $t("pages.fields.sort"),
    width: 96,
  },
  {
    key: "action",
    title: $t("pages.common.actions"),
    width: 160,
  },
]);

const linkColumns = computed(() => [
  {
    dataIndex: "name",
    title: $t("plugin.cms.fields.linkName"),
    width: 220,
  },
  {
    dataIndex: "url",
    title: $t("plugin.cms.fields.linkUrl"),
  },
  {
    dataIndex: "groupCode",
    title: $t("plugin.cms.fields.groupCode"),
    width: 120,
  },
  {
    dataIndex: "status",
    title: $t("pages.common.status"),
    width: 120,
  },
  {
    dataIndex: "sort",
    title: $t("pages.fields.sort"),
    width: 96,
  },
  {
    key: "action",
    title: $t("pages.common.actions"),
    width: 160,
  },
]);

onMounted(async () => {
  [
    statusDicts.value,
    categoryTypeDicts.value,
    articleStatusDicts.value,
    messageStatusDicts.value,
    yesNoDicts.value,
  ] = await Promise.all([
    dictStore.getDictOptionsAsync("cms_status"),
    dictStore.getDictOptionsAsync("cms_category_type"),
    dictStore.getDictOptionsAsync("cms_article_status"),
    dictStore.getDictOptionsAsync("cms_message_status"),
    dictStore.getDictOptionsAsync("cms_yes_no"),
  ]);
  await refreshAll();
});

watch(activeTab, async (tab) => {
  if (tab === "site") {
    await loadSite();
    return;
  }
  if (tab === "categories") {
    await loadCategories();
    return;
  }
  if (tab === "articles") {
    await Promise.all([loadCategories(), loadArticles()]);
    return;
  }
  if (tab === "slides") {
    await loadSlides();
    return;
  }
  if (tab === "links") {
    await loadLinks();
    return;
  }
  if (tab === "messages") {
    await loadMessages();
  }
});

watch(
  () => categoryForm.type,
  (type) => {
    if (type === CategoryTypeSingle && categoryForm.contentTemplate === "detail.html") {
      categoryForm.contentTemplate = "single.html";
    }
    if (type === CategoryTypeList && !categoryForm.listTemplate) {
      categoryForm.listTemplate = "list.html";
    }
    if (type === CategoryTypeList && categoryForm.contentTemplate === "single.html") {
      categoryForm.contentTemplate = "detail.html";
    }
  },
);

function toNumberOptions(items: any[]) {
  return items.map((item: any) => ({
    label: item.label,
    value: Number(item.value),
  }));
}

function countCategories(
  categories: Category[],
  predicate: (category: Category) => boolean = () => true,
) {
  let total = 0;
  for (const category of categories) {
    if (predicate(category)) {
      total += 1;
    }
    if (category.children?.length) {
      total += countCategories(category.children, predicate);
    }
  }
  return total;
}

function countCategoriesByType(categories: Category[], type: number) {
  return countCategories(categories, (category) => category.type === type);
}

function flattenCategories(
  categories: Category[],
  prefix = "",
  disabledRootId?: number,
) {
  const items: SelectNumberOption[] = [
    { label: $t("plugin.cms.placeholders.rootCategory"), value: 0 },
  ];

  function visit(
    nodes: Category[],
    parentPath: string,
    disabledAncestor = false,
  ) {
    for (const node of nodes) {
      const label = parentPath ? `${parentPath} / ${node.name}` : node.name;
      const disabled = disabledAncestor || node.id === disabledRootId;
      items.push({ disabled, label, value: node.id });
      if (node.children?.length) {
        visit(node.children, label, disabled);
      }
    }
  }

  visit(categories, prefix);
  return items;
}

function filterArticleCategoryTree(categories: Category[]): Category[] {
  const items: Category[] = [];
  for (const category of categories) {
    const children = filterArticleCategoryTree(category.children ?? []);
    if (category.type !== CategoryTypeExternal || children.length > 0) {
      items.push({
        ...category,
        children,
      });
    }
  }
  return items;
}

function flattenVisibleArticleCategoryNavItems(
  categories: Category[],
  expandedIds: Set<number>,
  depth = 0,
) {
  const items: {
    category: Category;
    depth: number;
    expanded: boolean;
    hasChildren: boolean;
  }[] = [];
  for (const category of categories) {
    const hasChildren = hasArticleCategoryChildren(category);
    const expanded = expandedIds.has(category.id);
    items.push({ category, depth, expanded, hasChildren });
    if (hasChildren && expanded) {
      items.push(
        ...flattenVisibleArticleCategoryNavItems(
          category.children ?? [],
          expandedIds,
          depth + 1,
        ),
      );
    }
  }
  return items;
}

function hasArticleCategoryChildren(category: Category) {
  return (category.children ?? []).length > 0;
}

function expandPathToArticleCategory(categoryId?: number) {
  if (!categoryId) {
    return;
  }
  const path = findCategoryPath(categoryRows.value, categoryId);
  if (path.length <= 1) {
    return;
  }
  expandedArticleCategoryIds.value = new Set(path.slice(0, -1));
}

function toggleArticleCategory(category: Category) {
  const next = new Set(expandedArticleCategoryIds.value);
  if (next.has(category.id)) {
    next.delete(category.id);
  } else {
    next.add(category.id);
  }
  expandedArticleCategoryIds.value = next;
}

function findCategoryById(categories: Category[], id?: number): Category | null {
  if (!id) {
    return null;
  }
  for (const category of categories) {
    if (category.id === id) {
      return category;
    }
    const child = findCategoryById(category.children ?? [], id);
    if (child) {
      return child;
    }
  }
  return null;
}

function findCategoryPath(categories: Category[], id: number, path: number[] = []) {
  for (const category of categories) {
    const nextPath = [...path, category.id];
    if (category.id === id) {
      return nextPath;
    }
    const childPath = findCategoryPath(category.children ?? [], id, nextPath);
    if (childPath.length > 0) {
      return childPath;
    }
  }
  return [];
}

function firstArticleCategoryByType(type?: number) {
  const flat = articleCategoryOptions.value;
  if (!type) {
    return flat[0]?.value;
  }
  const category = findFirstCategoryByType(categoryRows.value, type);
  return category?.id ?? flat[0]?.value;
}

function findFirstCategoryByType(categories: Category[], type: number): Category | null {
  for (const category of categories) {
    if (category.type === type) {
      return category;
    }
    const child = findFirstCategoryByType(category.children ?? [], type);
    if (child) {
      return child;
    }
  }
  return null;
}

async function selectArticleModel(categoryType: number, key: string) {
  selectedArticleSection.value = key;
  Object.assign(articleQuery, {
    categoryId: undefined,
    categoryType,
    includeChildren: true,
    pageNum: 1,
  });
  await loadArticles();
}

async function selectArticleCategory(category: Category) {
  selectedArticleSection.value = `category:${category.id}`;
  expandPathToArticleCategory(category.id);
  Object.assign(articleQuery, {
    categoryId: category.id,
    categoryType: undefined,
    includeChildren: true,
    pageNum: 1,
  });
  await loadArticles();
}

function selectArticleFilterCategory(value?: number) {
  if (value) {
    expandPathToArticleCategory(value);
  }
  Object.assign(articleQuery, {
    categoryType: value ? undefined : CategoryTypeList,
    includeChildren: true,
    pageNum: 1,
  });
  selectedArticleSection.value = value ? `category:${value}` : "model:list";
}

function formatDateTime(value?: string) {
  if (!value) {
    return "--";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleString();
}

function truncateText(value?: string, maxLength = 32) {
  if (!value) {
    return "--";
  }
  return value.length > maxLength ? `${value.slice(0, maxLength)}...` : value;
}

async function refreshAll() {
  if (!articleQuery.categoryType && !articleQuery.categoryId) {
    articleQuery.categoryType = CategoryTypeList;
  }
  await Promise.all([
    loadSite(),
    loadCategories(),
    loadArticles(),
    loadMessages(),
    loadSlides(),
    loadLinks(),
    loadDashboardMetrics(),
  ]);
}

async function loadDashboardMetrics() {
  dashboardLoading.value = true;
  try {
    const [allArticles, published, draft, pending] = await Promise.all([
      cmsArticleList({ pageNum: 1, pageSize: 1 }),
      cmsArticleList({ pageNum: 1, pageSize: 1, status: 1 }),
      cmsArticleList({ pageNum: 1, pageSize: 1, status: 0 }),
      cmsMessageList({ pageNum: 1, pageSize: 1, status: 0 }),
    ]);
    allArticleTotal.value = allArticles.total;
    publishedArticleTotal.value = published.total;
    draftArticleTotal.value = draft.total;
    pendingMessageTotal.value = pending.total;
  } finally {
    dashboardLoading.value = false;
  }
}

async function loadSite() {
  const data = await cmsSite();
  Object.assign(siteForm, data);
}

async function saveSite() {
  await cmsSiteUpdate(siteForm);
  message.success($t("pages.common.updateSuccess"));
  await loadSite();
}

async function loadCategories() {
  categoryRows.value = await cmsCategoryList();
}

function resetCategoryForm() {
  Object.assign(categoryForm, {
    code: "",
    contentTemplate: "detail.html",
    cover: "",
    description: "",
    id: undefined,
    keywords: "",
    listTemplate: "list.html",
    name: "",
    outlink: "",
    parentId: 0,
    path: "",
    sort: 0,
    status: 1,
    title: "",
    type: 1,
  });
}

function openCreateCategory() {
  categoryModalMode.value = "create";
  resetCategoryForm();
  categoryModalOpen.value = true;
}

function openEditCategory(row: Category) {
  categoryModalMode.value = "update";
  resetCategoryForm();
  Object.assign(categoryForm, row);
  categoryModalOpen.value = true;
}

async function submitCategory() {
  if (categoryModalMode.value === "update" && categoryForm.id) {
    await cmsCategoryUpdate(categoryForm.id, categoryForm);
    message.success($t("pages.common.updateSuccess"));
  } else {
    await cmsCategoryCreate(categoryForm);
    message.success($t("pages.common.createSuccess"));
  }
  categoryModalOpen.value = false;
  await Promise.all([loadCategories(), loadDashboardMetrics()]);
  await loadArticles();
}

async function deleteCategory(row: Category) {
  await cmsCategoryDelete(row.id);
  message.success($t("pages.common.deleteSuccess"));
  await Promise.all([loadCategories(), loadDashboardMetrics()]);
}

async function loadArticles() {
  articleLoading.value = true;
  try {
    const params = {
      ...articleQuery,
      includeChildren:
        articleQuery.includeChildren && !!articleQuery.categoryId
          ? true
          : undefined,
    };
    const resp = await cmsArticleList(params);
    articleRows.value = resp.items;
    articleTotal.value = resp.total;
  } finally {
    articleLoading.value = false;
  }
}

function resetArticleForm() {
  Object.assign(articleForm, {
    author: "",
    categoryId:
      articleQuery.categoryId ||
      firstArticleCategoryByType(articleQuery.categoryType),
    content: "",
    cover: "",
    description: "",
    id: undefined,
    isRecommend: 0,
    isTop: 0,
    keywords: "",
    slug: "",
    sort: 0,
    source: "",
    status: 0,
    subtitle: "",
    summary: "",
    tags: "",
    title: "",
  });
}

function openCreateArticle() {
  articleModalMode.value = "create";
  resetArticleForm();
  articleModalOpen.value = true;
}

async function openEditArticle(row: Article) {
  articleModalMode.value = "update";
  resetArticleForm();
  const detail = await cmsArticleInfo(row.id);
  Object.assign(articleForm, detail);
  articleModalOpen.value = true;
}

async function submitArticle() {
  if (articleModalMode.value === "update" && articleForm.id) {
    await cmsArticleUpdate(articleForm.id, articleForm);
    message.success($t("pages.common.updateSuccess"));
  } else {
    await cmsArticleCreate(articleForm);
    message.success($t("pages.common.createSuccess"));
  }
  articleModalOpen.value = false;
  await Promise.all([loadArticles(), loadDashboardMetrics()]);
}

async function deleteArticle(row: Article) {
  await cmsArticleDelete(row.id);
  message.success($t("pages.common.deleteSuccess"));
  await Promise.all([loadArticles(), loadDashboardMetrics()]);
}

function resetArticleQuery() {
  Object.assign(articleQuery, {
    categoryId: undefined,
    categoryType: CategoryTypeList,
    includeChildren: true,
    pageNum: 1,
    pageSize: 10,
    status: undefined,
    title: "",
  });
  selectedArticleSection.value = "model:list";
  expandedArticleCategoryIds.value = new Set();
  loadArticles();
}

async function loadMessages() {
  messageLoading.value = true;
  try {
    const resp = await cmsMessageList(messageQuery);
    messageRows.value = resp.items;
    messageTotal.value = resp.total;
  } finally {
    messageLoading.value = false;
  }
}

function resetMessageQuery() {
  Object.assign(messageQuery, {
    keyword: "",
    pageNum: 1,
    pageSize: 10,
    status: undefined,
  });
  loadMessages();
}

function openMessageModeration(row: Message, status = row.status) {
  currentMessage.value = row;
  Object.assign(messageForm, {
    reply: row.reply,
    status,
  });
  messageModalOpen.value = true;
}

async function submitMessageModeration() {
  if (!currentMessage.value) {
    return;
  }
  await cmsMessageUpdate(currentMessage.value.id, {
    reply: messageForm.reply,
    status: messageForm.status,
  });
  message.success($t("pages.common.updateSuccess"));
  messageModalOpen.value = false;
  await Promise.all([loadMessages(), loadDashboardMetrics()]);
}

async function deleteMessage(row: Message) {
  await cmsMessageDelete(row.id);
  message.success($t("pages.common.deleteSuccess"));
  await Promise.all([loadMessages(), loadDashboardMetrics()]);
}

async function loadSlides() {
  slideLoading.value = true;
  try {
    const resp = await cmsSlideList(slideQuery);
    slideRows.value = resp.items;
    slideTotal.value = resp.total;
  } finally {
    slideLoading.value = false;
  }
}

function resetSlideQuery() {
  Object.assign(slideQuery, {
    groupCode: "",
    keyword: "",
    pageNum: 1,
    pageSize: 10,
    status: undefined,
  });
  loadSlides();
}

function resetSlideForm() {
  Object.assign(slideForm, {
    groupCode: "1",
    id: undefined,
    image: "",
    link: "",
    sort: 0,
    status: 1,
    subtitle: "",
    title: "",
  });
}

function openCreateSlide() {
  slideModalMode.value = "create";
  resetSlideForm();
  slideModalOpen.value = true;
}

function openEditSlide(row: Slide) {
  slideModalMode.value = "update";
  resetSlideForm();
  Object.assign(slideForm, row);
  slideModalOpen.value = true;
}

async function submitSlide() {
  if (slideModalMode.value === "update" && slideForm.id) {
    await cmsSlideUpdate(slideForm.id, slideForm);
    message.success($t("pages.common.updateSuccess"));
  } else {
    await cmsSlideCreate(slideForm);
    message.success($t("pages.common.createSuccess"));
  }
  slideModalOpen.value = false;
  await loadSlides();
}

async function deleteSlide(row: Slide) {
  await cmsSlideDelete(row.id);
  message.success($t("pages.common.deleteSuccess"));
  await loadSlides();
}

async function loadLinks() {
  linkLoading.value = true;
  try {
    const resp = await cmsLinkList(linkQuery);
    linkRows.value = resp.items;
    linkTotal.value = resp.total;
  } finally {
    linkLoading.value = false;
  }
}

function resetLinkQuery() {
  Object.assign(linkQuery, {
    groupCode: "",
    keyword: "",
    pageNum: 1,
    pageSize: 10,
    status: undefined,
  });
  loadLinks();
}

function resetLinkForm() {
  Object.assign(linkForm, {
    groupCode: "1",
    id: undefined,
    logo: "",
    name: "",
    sort: 0,
    status: 1,
    url: "",
  });
}

function openCreateLink() {
  linkModalMode.value = "create";
  resetLinkForm();
  linkModalOpen.value = true;
}

function openEditLink(row: Link) {
  linkModalMode.value = "update";
  resetLinkForm();
  Object.assign(linkForm, row);
  linkModalOpen.value = true;
}

async function submitLink() {
  if (linkModalMode.value === "update" && linkForm.id) {
    await cmsLinkUpdate(linkForm.id, linkForm);
    message.success($t("pages.common.updateSuccess"));
  } else {
    await cmsLinkCreate(linkForm);
    message.success($t("pages.common.createSuccess"));
  }
  linkModalOpen.value = false;
  await loadLinks();
}

async function deleteLink(row: Link) {
  await cmsLinkDelete(row.id);
  message.success($t("pages.common.deleteSuccess"));
  await loadLinks();
}
</script>

<template>
  <Page :auto-content-height="true">
    <div class="cms-workbench" data-testid="cms-workbench">
      <section class="cms-overview" data-testid="cms-overview">
        <div class="cms-site-identity">
          <div class="cms-kicker">{{ $t("plugin.cms.overview.kicker") }}</div>
          <div class="cms-site-title-row">
            <h1>{{ siteForm.name || $t("plugin.cms.name") }}</h1>
            <a-tag :color="siteStatusTone">
              <DictTag :dicts="statusDicts" :value="String(siteForm.status)" />
            </a-tag>
          </div>
          <p>{{ siteForm.slogan || siteForm.description || "--" }}</p>
          <div class="cms-site-meta">
            <span>
              <span class="cms-meta-dot" />
              {{ siteDomainLabel }}
            </span>
            <span>
              <span class="cms-meta-dot" />
              {{
                $t("plugin.cms.metrics.draftArticles", {
                  count: draftArticleTotal,
                })
              }}
            </span>
            <span>
              <span class="cms-meta-dot" />
              {{
                $t("plugin.cms.metrics.enabledCategories", {
                  count: enabledCategoryTotal,
                })
              }}
            </span>
          </div>
        </div>

        <div class="cms-metric-grid" :aria-busy="dashboardLoading">
          <div
            v-for="item in metrics"
            :key="item.testId"
            class="cms-metric"
            :class="`cms-metric--${item.tone}`"
            :data-testid="item.testId"
          >
            <div class="cms-metric-icon">
              {{ item.icon }}
            </div>
            <div>
              <div class="cms-metric-value">{{ item.value }}</div>
              <div class="cms-metric-label">{{ item.label }}</div>
            </div>
          </div>
        </div>
      </section>

      <div class="cms-quickbar">
        <div class="cms-section-tabs">
          <button
            :class="{ 'is-active': activeTab === 'site' }"
            data-testid="cms-section-site"
            type="button"
            @click="activeTab = 'site'"
          >
            {{ $t("plugin.cms.tabs.site") }}
          </button>
          <button
            :class="{ 'is-active': activeTab === 'categories' }"
            data-testid="cms-section-categories"
            type="button"
            @click="activeTab = 'categories'"
          >
            {{ $t("plugin.cms.tabs.categories") }}
          </button>
          <button
            :class="{ 'is-active': activeTab === 'articles' }"
            data-testid="cms-section-articles"
            type="button"
            @click="activeTab = 'articles'"
          >
            {{ $t("plugin.cms.tabs.articles") }}
          </button>
          <button
            :class="{ 'is-active': activeTab === 'slides' }"
            data-testid="cms-section-slides"
            type="button"
            @click="activeTab = 'slides'"
          >
            {{ $t("plugin.cms.tabs.slides") }}
          </button>
          <button
            :class="{ 'is-active': activeTab === 'links' }"
            data-testid="cms-section-links"
            type="button"
            @click="activeTab = 'links'"
          >
            {{ $t("plugin.cms.tabs.links") }}
          </button>
          <button
            :class="{ 'is-active': activeTab === 'messages' }"
            data-testid="cms-section-messages"
            type="button"
            @click="activeTab = 'messages'"
          >
            {{ $t("plugin.cms.tabs.messages") }}
          </button>
        </div>
        <Space wrap>
          <a-button data-testid="cms-refresh-all" @click="refreshAll">
            {{ $t("plugin.cms.actions.refresh") }}
          </a-button>
          <a-button
            v-if="activeTab === 'categories'"
            data-testid="cms-category-add"
            type="primary"
            @click="openCreateCategory"
          >
            {{ $t("plugin.cms.actions.newCategory") }}
          </a-button>
          <a-button
            v-if="activeTab === 'articles'"
            data-testid="cms-article-add"
            type="primary"
            @click="openCreateArticle"
          >
            {{ $t("plugin.cms.actions.newArticle") }}
          </a-button>
          <a-button
            v-if="activeTab === 'slides'"
            data-testid="cms-slide-add"
            type="primary"
            @click="openCreateSlide"
          >
            {{ $t("plugin.cms.actions.newSlide") }}
          </a-button>
          <a-button
            v-if="activeTab === 'links'"
            data-testid="cms-link-add"
            type="primary"
            @click="openCreateLink"
          >
            {{ $t("plugin.cms.actions.newLink") }}
          </a-button>
        </Space>
      </div>

      <a-tabs v-model:active-key="activeTab" data-testid="cms-plugin-tabs">
        <a-tab-pane key="site" :tab="$t('plugin.cms.tabs.site')">
          <div class="cms-site-layout">
            <section class="cms-panel cms-panel--wide">
              <div class="cms-panel-head">
                <div>
                  <h2>{{ $t("plugin.cms.sections.siteProfile") }}</h2>
                  <p>{{ $t("plugin.cms.sections.siteProfileSubtitle") }}</p>
                </div>
                <Space>
                  <a-button data-testid="cms-site-refresh" @click="loadSite">
                    {{ $t("plugin.cms.actions.refresh") }}
                  </a-button>
                  <a-button
                    data-testid="cms-site-save"
                    type="primary"
                    @click="saveSite"
                  >
                    {{ $t("plugin.cms.actions.save") }}
                  </a-button>
                </Space>
              </div>
              <a-form
                :model="siteForm"
                layout="vertical"
                class="cms-form-grid"
              >
                <a-form-item
                  :label="$t('plugin.cms.fields.siteName')"
                  required
                >
                  <a-input v-model:value="siteForm.name" />
                </a-form-item>
                <a-form-item :label="$t('plugin.cms.fields.domain')">
                  <a-input v-model:value="siteForm.domain" />
                </a-form-item>
                <a-form-item :label="$t('pages.common.status')">
                  <a-select
                    v-model:value="siteForm.status"
                    :options="siteStatusOptions"
                  />
                </a-form-item>
                <div class="cms-site-media-grid cms-span-all">
                  <a-form-item :label="$t('plugin.cms.fields.logo')">
                    <div data-testid="cms-site-logo-upload">
                      <CmsImageUpload
                        v-model:value="siteForm.logo"
                        scene="cms-site"
                      />
                    </div>
                  </a-form-item>
                  <a-form-item :label="$t('plugin.cms.fields.weixin')">
                    <div data-testid="cms-site-weixin-upload">
                      <CmsImageUpload
                        v-model:value="siteForm.weixin"
                        scene="cms-site"
                      />
                    </div>
                  </a-form-item>
                </div>
                <a-form-item :label="$t('plugin.cms.fields.slogan')">
                  <a-input v-model:value="siteForm.slogan" />
                </a-form-item>
                <a-form-item :label="$t('plugin.cms.fields.keywords')">
                  <a-input v-model:value="siteForm.keywords" />
                </a-form-item>
                <a-form-item
                  :label="$t('plugin.cms.fields.description')"
                  class="cms-span-all"
                >
                  <a-textarea v-model:value="siteForm.description" :rows="3" />
                </a-form-item>
              </a-form>
            </section>

            <section class="cms-panel">
              <div class="cms-panel-head">
                <div>
                  <h2>{{ $t("plugin.cms.sections.contact") }}</h2>
                  <p>{{ $t("plugin.cms.sections.contactSubtitle") }}</p>
                </div>
              </div>
              <a-form :model="siteForm" layout="vertical">
                <a-form-item :label="$t('plugin.cms.fields.contact')">
                  <a-input v-model:value="siteForm.contact" />
                </a-form-item>
                <a-form-item :label="$t('plugin.cms.fields.phone')">
                  <a-input v-model:value="siteForm.phone" />
                </a-form-item>
                <a-form-item :label="$t('pages.fields.email')">
                  <a-input v-model:value="siteForm.email" />
                </a-form-item>
                <a-form-item :label="$t('plugin.cms.fields.address')">
                  <a-input v-model:value="siteForm.address" />
                </a-form-item>
                <a-form-item :label="$t('plugin.cms.fields.icp')">
                  <a-input v-model:value="siteForm.icp" />
                </a-form-item>
              </a-form>
            </section>
          </div>
        </a-tab-pane>

        <a-tab-pane key="categories" :tab="$t('plugin.cms.tabs.categories')">
          <section class="cms-panel">
            <div class="cms-panel-head">
              <div>
                <h2>{{ $t("plugin.cms.sections.categoryTree") }}</h2>
                <p>
                  {{
                    $t("plugin.cms.metrics.categorySummary", {
                      enabled: enabledCategoryTotal,
                      total: categoryTotal,
                    })
                  }}
                </p>
              </div>
              <a-button
                data-testid="cms-category-add-secondary"
                type="primary"
                @click="openCreateCategory"
              >
                {{ $t("plugin.cms.actions.newCategory") }}
              </a-button>
            </div>
            <a-table
              :columns="categoryColumns"
              :data-source="categoryRows"
              :pagination="false"
              data-testid="cms-category-table"
              row-key="id"
              size="middle"
            >
              <template #emptyText>
                <a-empty :description="$t('plugin.cms.empty.categories')" />
              </template>
              <template #bodyCell="{ column, record }">
                <template v-if="column.dataIndex === 'name'">
                  <div class="cms-primary-cell">
                    <strong>{{ record.name }}</strong>
                    <span>{{ record.path || record.outlink || "--" }}</span>
                    <span class="cms-template-line">
                      {{
                        [
                          record.listTemplate || "list.html",
                          record.contentTemplate || "detail.html",
                        ].join(" / ")
                      }}
                    </span>
                  </div>
                </template>
                <template v-else-if="column.dataIndex === 'type'">
                  <DictTag
                    :dicts="categoryTypeDicts"
                    :value="String(record.type)"
                  />
                </template>
                <template v-else-if="column.dataIndex === 'status'">
                  <DictTag :dicts="statusDicts" :value="String(record.status)" />
                </template>
                <template v-else-if="column.key === 'action'">
                  <Space>
                    <a-button
                      :data-testid="`cms-category-edit-${record.id}`"
                      type="link"
                      @click="openEditCategory(record)"
                    >
                      {{ $t("pages.common.edit") }}
                    </a-button>
                    <Popconfirm
                      :title="$t('pages.common.deleteConfirm')"
                      @confirm="deleteCategory(record)"
                    >
                      <a-button
                        :data-testid="`cms-category-delete-${record.id}`"
                        danger
                        type="link"
                      >
                        {{ $t("pages.common.delete") }}
                      </a-button>
                    </Popconfirm>
                  </Space>
                </template>
              </template>
            </a-table>
          </section>
        </a-tab-pane>

        <a-tab-pane key="articles" :tab="$t('plugin.cms.tabs.articles')">
          <div class="cms-content-layout">
            <aside class="cms-panel cms-content-nav" data-testid="cms-content-nav">
              <div class="cms-content-nav-head">
                <h2>{{ $t("plugin.cms.sections.contentModules") }}</h2>
                <p>{{ $t("plugin.cms.sections.contentModulesSubtitle") }}</p>
              </div>
              <div class="cms-content-models">
                <button
                  v-for="model in articleSectionGroups"
                  :key="model.key"
                  class="cms-content-model"
                  :class="[
                    `cms-content-model--${model.tone}`,
                    { 'is-active': selectedArticleSection === model.key },
                  ]"
                  :data-testid="`cms-content-model-${model.key.replace(':', '-')}`"
                  type="button"
                  @click="selectArticleModel(model.categoryType, model.key)"
                >
                  <span>{{ model.label }}</span>
                  <strong>{{ model.total }}</strong>
                </button>
              </div>
              <div class="cms-content-category-tree">
                <div
                  v-for="item in articleCategoryNavItems"
                  :key="item.category.id"
                  class="cms-content-category-row"
                  :class="{
                    'cms-content-category-row--child': item.depth > 0,
                    'is-active':
                      selectedArticleSection === `category:${item.category.id}`,
                  }"
                  :data-testid="`cms-content-category-row-${item.category.id}`"
                  :style="{ paddingLeft: `${10 + item.depth * 14}px` }"
                >
                  <button
                    v-if="item.hasChildren"
                    class="cms-content-category-toggle"
                    :aria-expanded="item.expanded"
                    :data-testid="`cms-content-category-toggle-${item.category.id}`"
                    type="button"
                    @click.stop="toggleArticleCategory(item.category)"
                  >
                    {{ item.expanded ? "-" : "+" }}
                  </button>
                  <span v-else class="cms-content-category-spacer"></span>
                  <button
                    class="cms-content-category"
                    :data-testid="`cms-content-category-${item.category.id}`"
                    type="button"
                    @click="selectArticleCategory(item.category)"
                  >
                    <span>{{ item.category.name }}</span>
                    <DictTag
                      :dicts="categoryTypeDicts"
                      :value="String(item.category.type)"
                    />
                  </button>
                </div>
              </div>
            </aside>

            <section class="cms-panel">
              <div class="cms-panel-head">
                <div>
                  <h2>{{ selectedArticleSectionLabel }}</h2>
                  <p>
                    {{
                      $t("plugin.cms.metrics.articleSummary", {
                        draft: draftArticleTotal,
                        published: publishedArticleTotal,
                      })
                    }}
                  </p>
                </div>
                <a-button
                  data-testid="cms-article-add-secondary"
                  type="primary"
                  @click="openCreateArticle"
                >
                  {{ $t("plugin.cms.actions.newArticle") }}
                </a-button>
              </div>
              <div class="cms-filterbar cms-article-filterbar">
                <a-input
                  v-model:value="articleQuery.title"
                  :placeholder="$t('plugin.cms.placeholders.articleTitle')"
                  class="cms-filter-input"
                  data-testid="cms-article-title-filter"
                  allow-clear
                />
                <a-select
                  v-model:value="articleQuery.categoryId"
                  :options="articleCategoryOptions"
                  :placeholder="$t('plugin.cms.placeholders.category')"
                  class="cms-filter-select"
                  allow-clear
                  @change="selectArticleFilterCategory"
                />
                <a-select
                  v-model:value="articleQuery.status"
                  :options="articleStatusOptions"
                  :placeholder="$t('plugin.cms.placeholders.status')"
                  class="cms-filter-select"
                  allow-clear
                />
                <Space>
                  <a-button data-testid="cms-article-query" @click="loadArticles">
                    {{ $t("plugin.cms.actions.query") }}
                  </a-button>
                  <a-button @click="resetArticleQuery">
                    {{ $t("plugin.cms.actions.reset") }}
                  </a-button>
                </Space>
              </div>
              <a-table
                :columns="articleColumns"
                :data-source="articleRows"
                :loading="articleLoading"
                class="cms-article-table"
                data-testid="cms-article-table"
                :pagination="{
                  current: articleQuery.pageNum,
                  pageSize: articleQuery.pageSize,
                  total: articleTotal,
                  showSizeChanger: true,
                  onChange: (page: number, pageSize: number) => {
                    articleQuery.pageNum = page;
                    articleQuery.pageSize = pageSize;
                    loadArticles();
                  },
                }"
                row-key="id"
                size="middle"
                table-layout="fixed"
              >
                <template #emptyText>
                  <a-empty :description="$t('plugin.cms.empty.articles')" />
                </template>
                <template #bodyCell="{ column, record }">
                  <template v-if="column.dataIndex === 'title'">
                    <div class="cms-primary-cell cms-article-title-cell">
                      <strong
                        class="cms-article-title-text"
                        :title="record.title"
                      >
                        {{ truncateText(record.title, 32) }}
                      </strong>
                      <span :title="record.summary || record.slug">
                        {{ truncateText(record.summary || record.slug, 38) }}
                      </span>
                    </div>
                  </template>
                  <template v-else-if="column.dataIndex === 'categoryName'">
                    <span class="cms-table-ellipsis" :title="record.categoryName">
                      {{ record.categoryName || "--" }}
                    </span>
                  </template>
                  <template v-else-if="column.dataIndex === 'status'">
                    <DictTag
                      :dicts="articleStatusDicts"
                      :value="String(record.status)"
                    />
                  </template>
                  <template v-else-if="column.dataIndex === 'isRecommend'">
                    <DictTag
                      :dicts="yesNoDicts"
                      :value="String(record.isRecommend)"
                    />
                  </template>
                  <template v-else-if="column.key === 'action'">
                    <Space>
                      <a-button
                        :data-testid="`cms-article-edit-${record.id}`"
                        type="link"
                        @click="openEditArticle(record)"
                      >
                        {{ $t("pages.common.edit") }}
                      </a-button>
                      <Popconfirm
                        :title="$t('pages.common.deleteConfirm')"
                        @confirm="deleteArticle(record)"
                      >
                        <a-button
                          :data-testid="`cms-article-delete-${record.id}`"
                          danger
                          type="link"
                        >
                          {{ $t("pages.common.delete") }}
                        </a-button>
                      </Popconfirm>
                    </Space>
                  </template>
                </template>
              </a-table>
            </section>
          </div>
        </a-tab-pane>

        <a-tab-pane key="slides" :tab="$t('plugin.cms.tabs.slides')">
          <section class="cms-panel">
            <div class="cms-panel-head">
              <div>
                <h2>{{ $t("plugin.cms.sections.slideManager") }}</h2>
                <p>{{ $t("plugin.cms.sections.slideManagerSubtitle") }}</p>
              </div>
              <a-button
                data-testid="cms-slide-add-secondary"
                type="primary"
                @click="openCreateSlide"
              >
                {{ $t("plugin.cms.actions.newSlide") }}
              </a-button>
            </div>
            <div class="cms-filterbar">
              <a-input
                v-model:value="slideQuery.keyword"
                :placeholder="$t('plugin.cms.placeholders.slideKeyword')"
                class="cms-filter-input"
                data-testid="cms-slide-keyword-filter"
                allow-clear
              />
              <a-input
                v-model:value="slideQuery.groupCode"
                :placeholder="$t('plugin.cms.placeholders.groupCode')"
                class="cms-filter-select"
                data-testid="cms-slide-group-filter"
                allow-clear
              />
              <a-select
                v-model:value="slideQuery.status"
                :options="siteStatusOptions"
                :placeholder="$t('plugin.cms.placeholders.status')"
                class="cms-filter-select"
                allow-clear
              />
              <Space>
                <a-button data-testid="cms-slide-query" @click="loadSlides">
                  {{ $t("plugin.cms.actions.query") }}
                </a-button>
                <a-button @click="resetSlideQuery">
                  {{ $t("plugin.cms.actions.reset") }}
                </a-button>
              </Space>
            </div>
            <a-table
              :columns="slideColumns"
              :data-source="slideRows"
              :loading="slideLoading"
              data-testid="cms-slide-table"
              :pagination="{
                current: slideQuery.pageNum,
                pageSize: slideQuery.pageSize,
                total: slideTotal,
                showSizeChanger: true,
                onChange: (page: number, pageSize: number) => {
                  slideQuery.pageNum = page;
                  slideQuery.pageSize = pageSize;
                  loadSlides();
                },
              }"
              row-key="id"
              size="middle"
            >
              <template #emptyText>
                <a-empty :description="$t('plugin.cms.empty.slides')" />
              </template>
              <template #bodyCell="{ column, record }">
                <template v-if="column.dataIndex === 'title'">
                  <div class="cms-primary-cell">
                    <strong>{{ record.title }}</strong>
                    <span>{{ record.subtitle || record.link || "--" }}</span>
                  </div>
                </template>
                <template v-else-if="column.dataIndex === 'image'">
                  <span class="cms-table-ellipsis" :title="record.image">
                    {{ truncateText(record.image, 30) }}
                  </span>
                </template>
                <template v-else-if="column.dataIndex === 'status'">
                  <DictTag :dicts="statusDicts" :value="String(record.status)" />
                </template>
                <template v-else-if="column.key === 'action'">
                  <Space>
                    <a-button
                      :data-testid="`cms-slide-edit-${record.id}`"
                      type="link"
                      @click="openEditSlide(record)"
                    >
                      {{ $t("pages.common.edit") }}
                    </a-button>
                    <Popconfirm
                      :title="$t('pages.common.deleteConfirm')"
                      @confirm="deleteSlide(record)"
                    >
                      <a-button
                        :data-testid="`cms-slide-delete-${record.id}`"
                        danger
                        type="link"
                      >
                        {{ $t("pages.common.delete") }}
                      </a-button>
                    </Popconfirm>
                  </Space>
                </template>
              </template>
            </a-table>
          </section>
        </a-tab-pane>

        <a-tab-pane key="links" :tab="$t('plugin.cms.tabs.links')">
          <section class="cms-panel">
            <div class="cms-panel-head">
              <div>
                <h2>{{ $t("plugin.cms.sections.linkManager") }}</h2>
                <p>{{ $t("plugin.cms.sections.linkManagerSubtitle") }}</p>
              </div>
              <a-button
                data-testid="cms-link-add-secondary"
                type="primary"
                @click="openCreateLink"
              >
                {{ $t("plugin.cms.actions.newLink") }}
              </a-button>
            </div>
            <div class="cms-filterbar">
              <a-input
                v-model:value="linkQuery.keyword"
                :placeholder="$t('plugin.cms.placeholders.linkKeyword')"
                class="cms-filter-input"
                data-testid="cms-link-keyword-filter"
                allow-clear
              />
              <a-input
                v-model:value="linkQuery.groupCode"
                :placeholder="$t('plugin.cms.placeholders.groupCode')"
                class="cms-filter-select"
                data-testid="cms-link-group-filter"
                allow-clear
              />
              <a-select
                v-model:value="linkQuery.status"
                :options="siteStatusOptions"
                :placeholder="$t('plugin.cms.placeholders.status')"
                class="cms-filter-select"
                allow-clear
              />
              <Space>
                <a-button data-testid="cms-link-query" @click="loadLinks">
                  {{ $t("plugin.cms.actions.query") }}
                </a-button>
                <a-button @click="resetLinkQuery">
                  {{ $t("plugin.cms.actions.reset") }}
                </a-button>
              </Space>
            </div>
            <a-table
              :columns="linkColumns"
              :data-source="linkRows"
              :loading="linkLoading"
              data-testid="cms-link-table"
              :pagination="{
                current: linkQuery.pageNum,
                pageSize: linkQuery.pageSize,
                total: linkTotal,
                showSizeChanger: true,
                onChange: (page: number, pageSize: number) => {
                  linkQuery.pageNum = page;
                  linkQuery.pageSize = pageSize;
                  loadLinks();
                },
              }"
              row-key="id"
              size="middle"
            >
              <template #emptyText>
                <a-empty :description="$t('plugin.cms.empty.links')" />
              </template>
              <template #bodyCell="{ column, record }">
                <template v-if="column.dataIndex === 'name'">
                  <div class="cms-primary-cell">
                    <strong>{{ record.name }}</strong>
                    <span>{{ record.logo || "--" }}</span>
                  </div>
                </template>
                <template v-else-if="column.dataIndex === 'url'">
                  <a
                    class="cms-table-ellipsis"
                    :href="record.url"
                    :title="record.url"
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    {{ record.url }}
                  </a>
                </template>
                <template v-else-if="column.dataIndex === 'status'">
                  <DictTag :dicts="statusDicts" :value="String(record.status)" />
                </template>
                <template v-else-if="column.key === 'action'">
                  <Space>
                    <a-button
                      :data-testid="`cms-link-edit-${record.id}`"
                      type="link"
                      @click="openEditLink(record)"
                    >
                      {{ $t("pages.common.edit") }}
                    </a-button>
                    <Popconfirm
                      :title="$t('pages.common.deleteConfirm')"
                      @confirm="deleteLink(record)"
                    >
                      <a-button
                        :data-testid="`cms-link-delete-${record.id}`"
                        danger
                        type="link"
                      >
                        {{ $t("pages.common.delete") }}
                      </a-button>
                    </Popconfirm>
                  </Space>
                </template>
              </template>
            </a-table>
          </section>
        </a-tab-pane>

        <a-tab-pane key="messages" :tab="$t('plugin.cms.tabs.messages')">
          <section class="cms-panel">
            <div class="cms-panel-head">
              <div>
                <h2>{{ $t("plugin.cms.sections.messageInbox") }}</h2>
                <p>
                  {{
                    $t("plugin.cms.metrics.messageSummary", {
                      pending: pendingMessageTotal,
                      total: messageTotal,
                    })
                  }}
                </p>
              </div>
            </div>
            <div class="cms-filterbar">
              <a-input
                v-model:value="messageQuery.keyword"
                :placeholder="$t('plugin.cms.placeholders.messageKeyword')"
                class="cms-filter-input"
                data-testid="cms-message-keyword-filter"
                allow-clear
              />
              <a-select
                v-model:value="messageQuery.status"
                :options="messageStatusOptions"
                :placeholder="$t('plugin.cms.placeholders.status')"
                class="cms-filter-select"
                allow-clear
              />
              <Space>
                <a-button data-testid="cms-message-query" @click="loadMessages">
                  {{ $t("plugin.cms.actions.query") }}
                </a-button>
                <a-button @click="resetMessageQuery">
                  {{ $t("plugin.cms.actions.reset") }}
                </a-button>
              </Space>
            </div>
            <div
              v-if="pendingMessages.length"
              class="cms-message-strip"
              data-testid="cms-pending-message-strip"
            >
              <article
                v-for="item in pendingMessages"
                :key="item.id"
                class="cms-message-card"
              >
                <div>
                  <strong>{{ item.name }}</strong>
                  <span>{{ formatDateTime(item.createdAt) }}</span>
                </div>
                <p>{{ item.content }}</p>
                <a-button
                  size="small"
                  type="link"
                  @click="openMessageModeration(item, 1)"
                >
                  {{ $t("plugin.cms.actions.review") }}
                </a-button>
              </article>
            </div>
            <a-table
              :columns="messageColumns"
              :data-source="messageRows"
              :loading="messageLoading"
              data-testid="cms-message-table"
              :pagination="{
                current: messageQuery.pageNum,
                pageSize: messageQuery.pageSize,
                total: messageTotal,
                showSizeChanger: true,
                onChange: (page: number, pageSize: number) => {
                  messageQuery.pageNum = page;
                  messageQuery.pageSize = pageSize;
                  loadMessages();
                },
              }"
              row-key="id"
              size="middle"
            >
              <template #emptyText>
                <a-empty :description="$t('plugin.cms.empty.messages')" />
              </template>
              <template #bodyCell="{ column, record }">
                <template v-if="column.dataIndex === 'content'">
                  <div class="cms-primary-cell">
                    <strong>{{ record.content }}</strong>
                    <span>{{ formatDateTime(record.createdAt) }}</span>
                  </div>
                </template>
                <template v-else-if="column.dataIndex === 'status'">
                  <DictTag
                    :dicts="messageStatusDicts"
                    :value="String(record.status)"
                  />
                </template>
                <template v-else-if="column.key === 'action'">
                  <Space>
                    <a-button
                      type="link"
                      :data-testid="`cms-message-approve-${record.id}`"
                      @click="openMessageModeration(record, 1)"
                    >
                      {{ $t("plugin.cms.actions.approve") }}
                    </a-button>
                    <a-button
                      danger
                      type="link"
                      :data-testid="`cms-message-reject-${record.id}`"
                      @click="openMessageModeration(record, 2)"
                    >
                      {{ $t("plugin.cms.actions.reject") }}
                    </a-button>
                    <Popconfirm
                      :title="$t('pages.common.deleteConfirm')"
                      @confirm="deleteMessage(record)"
                    >
                      <a-button
                        :data-testid="`cms-message-delete-${record.id}`"
                        danger
                        type="link"
                      >
                        {{ $t("pages.common.delete") }}
                      </a-button>
                    </Popconfirm>
                  </Space>
                </template>
              </template>
            </a-table>
          </section>
        </a-tab-pane>
      </a-tabs>
    </div>

    <a-modal
      v-model:open="categoryModalOpen"
      data-testid="cms-category-modal"
      :title="
        categoryModalMode === 'update'
          ? $t('plugin.cms.dialogs.editCategory')
          : $t('plugin.cms.dialogs.createCategory')
      "
      :width="720"
      @ok="submitCategory"
    >
      <a-form :model="categoryForm" layout="vertical" class="cms-form-grid">
        <a-form-item :label="$t('plugin.cms.fields.parentCategory')">
          <a-select
            v-model:value="categoryForm.parentId"
            :options="categoryOptions"
            data-testid="cms-category-parent-input"
          />
        </a-form-item>
        <a-form-item :label="$t('plugin.cms.fields.categoryName')" required>
          <a-input
            v-model:value="categoryForm.name"
            data-testid="cms-category-name-input"
          />
        </a-form-item>
        <a-form-item :label="$t('plugin.cms.fields.categoryCode')" required>
          <a-input
            v-model:value="categoryForm.code"
            data-testid="cms-category-code-input"
          />
        </a-form-item>
        <a-form-item :label="$t('plugin.cms.fields.categoryType')" required>
          <a-select
            v-model:value="categoryForm.type"
            :options="categoryTypeOptions"
            data-testid="cms-category-type-input"
          />
        </a-form-item>
        <a-form-item :label="$t('plugin.cms.fields.path')">
          <a-input
            v-model:value="categoryForm.path"
            data-testid="cms-category-path-input"
          />
        </a-form-item>
        <a-form-item :label="$t('plugin.cms.fields.listTemplate')">
          <a-select
            v-model:value="categoryForm.listTemplate"
            :options="categoryListTemplateOptions"
            data-testid="cms-category-list-template-input"
          />
        </a-form-item>
        <a-form-item :label="$t('plugin.cms.fields.contentTemplate')">
          <a-select
            v-model:value="categoryForm.contentTemplate"
            :options="categoryContentTemplateOptions"
            data-testid="cms-category-content-template-input"
          />
        </a-form-item>
        <a-form-item :label="$t('plugin.cms.fields.outlink')">
          <a-input v-model:value="categoryForm.outlink" />
        </a-form-item>
        <a-form-item :label="$t('pages.common.status')">
          <a-select
            v-model:value="categoryForm.status"
            :options="siteStatusOptions"
          />
        </a-form-item>
        <a-form-item :label="$t('pages.fields.sort')">
          <a-input-number
            v-model:value="categoryForm.sort"
            class="w-full"
            :min="0"
          />
        </a-form-item>
        <a-form-item
          :label="$t('plugin.cms.fields.description')"
          class="cms-span-all"
        >
          <a-textarea v-model:value="categoryForm.description" :rows="3" />
        </a-form-item>
      </a-form>
    </a-modal>

    <a-modal
      v-model:open="articleModalOpen"
      data-testid="cms-article-modal"
      class="cms-article-modal"
      :title="
        articleModalMode === 'update'
          ? $t('plugin.cms.dialogs.editArticle')
          : $t('plugin.cms.dialogs.createArticle')
      "
      :width="980"
      wrap-class-name="cms-article-modal-wrap"
      @ok="submitArticle"
    >
      <a-form
        :model="articleForm"
        layout="vertical"
        class="cms-form-grid cms-article-form"
      >
        <a-form-item :label="$t('plugin.cms.fields.articleTitle')" required>
          <a-input
            v-model:value="articleForm.title"
            data-testid="cms-article-title-input"
          />
        </a-form-item>
        <a-form-item :label="$t('plugin.cms.fields.slug')" required>
          <a-input
            v-model:value="articleForm.slug"
            data-testid="cms-article-slug-input"
          />
        </a-form-item>
        <a-form-item :label="$t('plugin.cms.fields.categoryName')" required>
          <a-select
            v-model:value="articleForm.categoryId"
            :options="articleCategoryOptions"
            data-testid="cms-article-category-input"
            show-search
          />
        </a-form-item>
        <a-form-item :label="$t('pages.common.status')">
          <a-select
            v-model:value="articleForm.status"
            :options="articleStatusOptions"
            data-testid="cms-article-status-input"
          />
        </a-form-item>
        <a-form-item :label="$t('plugin.cms.fields.subtitle')">
          <a-input v-model:value="articleForm.subtitle" />
        </a-form-item>
        <a-form-item :label="$t('plugin.cms.fields.author')">
          <a-input
            v-model:value="articleForm.author"
            data-testid="cms-article-author-input"
          />
        </a-form-item>
        <a-form-item :label="$t('plugin.cms.fields.source')">
          <a-input
            v-model:value="articleForm.source"
            data-testid="cms-article-source-input"
          />
        </a-form-item>
        <a-form-item :label="$t('plugin.cms.fields.tags')">
          <a-input
            v-model:value="articleForm.tags"
            data-testid="cms-article-tags-input"
          />
        </a-form-item>
        <a-form-item :label="$t('plugin.cms.fields.cover')">
          <div data-testid="cms-article-cover-upload">
            <CmsImageUpload
              v-model:value="articleForm.cover"
              scene="other"
              :max-count="1"
            />
          </div>
        </a-form-item>
        <a-form-item :label="$t('plugin.cms.fields.recommend')">
          <a-select
            v-model:value="articleForm.isRecommend"
            :options="yesNoOptions"
            data-testid="cms-article-recommend-input"
          />
        </a-form-item>
        <a-form-item :label="$t('plugin.cms.fields.top')">
          <a-select v-model:value="articleForm.isTop" :options="yesNoOptions" />
        </a-form-item>
        <a-form-item :label="$t('pages.fields.sort')">
          <a-input-number
            v-model:value="articleForm.sort"
            class="w-full"
            :min="0"
          />
        </a-form-item>
        <a-form-item
          :label="$t('plugin.cms.fields.summary')"
          class="cms-span-all"
        >
          <a-textarea
            v-model:value="articleForm.summary"
            :rows="2"
            data-testid="cms-article-summary-input"
          />
        </a-form-item>
        <a-form-item
          :label="$t('plugin.cms.fields.content')"
          class="cms-span-all"
          required
        >
          <div data-testid="cms-article-content-editor">
            <CmsRichTextEditor
              v-model="articleForm.content"
              :height="300"
              scene="other"
            />
          </div>
        </a-form-item>
      </a-form>
    </a-modal>

    <a-modal
      v-model:open="messageModalOpen"
      data-testid="cms-message-modal"
      :title="$t('plugin.cms.dialogs.reviewMessage')"
      :width="720"
      @ok="submitMessageModeration"
    >
      <div v-if="currentMessage" class="cms-message-detail">
        <div class="cms-message-detail-head">
          <strong>{{ currentMessage.name }}</strong>
          <span>{{ currentMessage.email || currentMessage.mobile || "--" }}</span>
          <span>{{ formatDateTime(currentMessage.createdAt) }}</span>
        </div>
        <p>{{ currentMessage.content }}</p>
      </div>
      <a-form :model="messageForm" layout="vertical">
        <a-form-item :label="$t('pages.common.status')">
          <a-select
            v-model:value="messageForm.status"
            :options="messageStatusOptions"
            data-testid="cms-message-status-input"
          />
        </a-form-item>
        <a-form-item :label="$t('plugin.cms.fields.reply')">
          <a-textarea
            v-model:value="messageForm.reply"
            :rows="4"
            data-testid="cms-message-reply-input"
          />
        </a-form-item>
      </a-form>
    </a-modal>

    <a-modal
      v-model:open="slideModalOpen"
      data-testid="cms-slide-modal"
      :title="
        slideModalMode === 'update'
          ? $t('plugin.cms.dialogs.editSlide')
          : $t('plugin.cms.dialogs.createSlide')
      "
      :width="720"
      @ok="submitSlide"
    >
      <a-form :model="slideForm" layout="vertical" class="cms-form-grid">
        <a-form-item :label="$t('plugin.cms.fields.slideTitle')" required>
          <a-input
            v-model:value="slideForm.title"
            data-testid="cms-slide-title-input"
          />
        </a-form-item>
        <a-form-item :label="$t('plugin.cms.fields.groupCode')">
          <a-input
            v-model:value="slideForm.groupCode"
            data-testid="cms-slide-group-input"
          />
        </a-form-item>
        <a-form-item :label="$t('plugin.cms.fields.subtitle')">
          <a-input v-model:value="slideForm.subtitle" />
        </a-form-item>
        <a-form-item :label="$t('plugin.cms.fields.slideLink')">
          <a-input v-model:value="slideForm.link" />
        </a-form-item>
        <a-form-item :label="$t('pages.common.status')">
          <a-select
            v-model:value="slideForm.status"
            :options="siteStatusOptions"
          />
        </a-form-item>
        <a-form-item :label="$t('pages.fields.sort')">
          <a-input-number
            v-model:value="slideForm.sort"
            class="w-full"
            :min="0"
          />
        </a-form-item>
        <a-form-item
          :label="$t('plugin.cms.fields.slideImage')"
          class="cms-span-all"
          required
        >
          <div data-testid="cms-slide-image-upload">
            <CmsImageUpload
              v-model:value="slideForm.image"
              scene="cms-site"
              :max-count="1"
            />
          </div>
        </a-form-item>
      </a-form>
    </a-modal>

    <a-modal
      v-model:open="linkModalOpen"
      data-testid="cms-link-modal"
      :title="
        linkModalMode === 'update'
          ? $t('plugin.cms.dialogs.editLink')
          : $t('plugin.cms.dialogs.createLink')
      "
      :width="720"
      @ok="submitLink"
    >
      <a-form :model="linkForm" layout="vertical" class="cms-form-grid">
        <a-form-item :label="$t('plugin.cms.fields.linkName')" required>
          <a-input
            v-model:value="linkForm.name"
            data-testid="cms-link-name-input"
          />
        </a-form-item>
        <a-form-item :label="$t('plugin.cms.fields.groupCode')">
          <a-input
            v-model:value="linkForm.groupCode"
            data-testid="cms-link-group-input"
          />
        </a-form-item>
        <a-form-item
          :label="$t('plugin.cms.fields.linkUrl')"
          class="cms-span-all"
          required
        >
          <a-input
            v-model:value="linkForm.url"
            data-testid="cms-link-url-input"
          />
        </a-form-item>
        <a-form-item :label="$t('pages.common.status')">
          <a-select
            v-model:value="linkForm.status"
            :options="siteStatusOptions"
          />
        </a-form-item>
        <a-form-item :label="$t('pages.fields.sort')">
          <a-input-number
            v-model:value="linkForm.sort"
            class="w-full"
            :min="0"
          />
        </a-form-item>
        <a-form-item :label="$t('plugin.cms.fields.linkLogo')" class="cms-span-all">
          <div data-testid="cms-link-logo-upload">
            <CmsImageUpload
              v-model:value="linkForm.logo"
              scene="cms-site"
              :max-count="1"
            />
          </div>
        </a-form-item>
      </a-form>
    </a-modal>
  </Page>
</template>

<style scoped>
.cms-workbench {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-height: calc(100vh - 152px);
}

.cms-overview {
  display: grid;
  grid-template-columns: minmax(0, 1.45fr) minmax(440px, 1fr);
  gap: 18px;
  padding: 20px;
  color: #f8fafc;
  background:
    linear-gradient(135deg, rgba(15, 23, 42, 0.96), rgba(17, 94, 89, 0.9)),
    radial-gradient(circle at 82% 8%, rgba(56, 189, 248, 0.34), transparent 28%);
  border: 1px solid rgba(148, 163, 184, 0.24);
  border-radius: 8px;
}

.cms-kicker {
  margin-bottom: 8px;
  font-size: 12px;
  font-weight: 700;
  color: #99f6e4;
  text-transform: uppercase;
}

.cms-site-title-row {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  align-items: center;
}

.cms-site-title-row h1 {
  max-width: 760px;
  margin: 0;
  overflow-wrap: anywhere;
  font-size: 26px;
  font-weight: 720;
  line-height: 1.25;
}

.cms-site-identity p {
  max-width: 760px;
  margin: 10px 0 0;
  color: #cbd5e1;
  overflow-wrap: anywhere;
}

.cms-site-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 18px;
}

.cms-site-meta > span {
  display: inline-flex;
  gap: 6px;
  align-items: center;
  min-height: 30px;
  padding: 4px 10px;
  color: #e2e8f0;
  background: rgba(15, 23, 42, 0.38);
  border: 1px solid rgba(226, 232, 240, 0.18);
  border-radius: 999px;
}

.cms-meta-dot {
  width: 6px;
  height: 6px;
  background: #5eead4;
  border-radius: 999px;
}

.cms-metric-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.cms-metric {
  display: grid;
  grid-template-columns: 42px minmax(0, 1fr);
  gap: 12px;
  align-items: center;
  min-height: 92px;
  padding: 14px;
  background: rgba(255, 255, 255, 0.92);
  border: 1px solid rgba(226, 232, 240, 0.8);
  border-radius: 8px;
}

.cms-metric-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 42px;
  height: 42px;
  font-size: 20px;
  border-radius: 8px;
}

.cms-metric-value {
  font-size: 24px;
  font-weight: 760;
  line-height: 1.1;
  color: #0f172a;
}

.cms-metric-label {
  margin-top: 5px;
  font-size: 12px;
  color: #64748b;
}

.cms-metric--teal .cms-metric-icon {
  color: #0f766e;
  background: #ccfbf1;
}

.cms-metric--blue .cms-metric-icon {
  color: #1d4ed8;
  background: #dbeafe;
}

.cms-metric--green .cms-metric-icon {
  color: #15803d;
  background: #dcfce7;
}

.cms-metric--amber .cms-metric-icon {
  color: #b45309;
  background: #fef3c7;
}

.cms-quickbar {
  display: flex;
  gap: 12px;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
}

.cms-section-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.cms-section-tabs button {
  min-height: 34px;
  padding: 0 13px;
  color: #475569;
  cursor: pointer;
  background: transparent;
  border: 1px solid transparent;
  border-radius: 7px;
}

.cms-section-tabs button.is-active {
  color: #0f766e;
  background: #f0fdfa;
  border-color: #99f6e4;
}

.cms-panel {
  padding: 18px;
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
}

.cms-panel-head {
  display: flex;
  gap: 14px;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 16px;
}

.cms-panel-head h2 {
  margin: 0;
  font-size: 17px;
  font-weight: 700;
  color: #111827;
}

.cms-panel-head p {
  margin: 4px 0 0;
  color: #64748b;
}

.cms-site-layout {
  display: grid;
  grid-template-columns: minmax(0, 1.45fr) minmax(320px, 0.55fr);
  gap: 16px;
}

.cms-form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  column-gap: 14px;
}

.cms-form-grid .cms-span-all {
  grid-column: 1 / -1;
}

.cms-site-media-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.cms-site-media-grid :deep(.ant-form-item) {
  margin-bottom: 0;
}

.cms-site-media-grid :deep(.ant-upload-list-picture-card) {
  margin-bottom: 4px;
}

.cms-filterbar {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  align-items: center;
  padding: 12px;
  margin-bottom: 14px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
}

.cms-filter-input {
  width: min(280px, 100%);
}

.cms-filter-select {
  width: min(220px, 100%);
}

.cms-article-filterbar {
  display: grid;
  grid-template-columns: minmax(220px, 1fr) minmax(180px, 0.8fr) minmax(160px, 0.7fr) auto;
}

.cms-article-filterbar .cms-filter-input,
.cms-article-filterbar .cms-filter-select {
  width: 100%;
}

.cms-content-layout {
  display: grid;
  grid-template-columns: 252px minmax(0, 1fr);
  gap: 16px;
  align-items: start;
}

.cms-content-nav {
  padding: 14px;
}

.cms-content-nav-head h2 {
  margin: 0;
  font-size: 16px;
  font-weight: 700;
  color: #111827;
}

.cms-content-nav-head p {
  margin: 4px 0 12px;
  font-size: 12px;
  color: #64748b;
}

.cms-content-models,
.cms-content-category-tree {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.cms-content-model {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 42px;
  padding: 8px 10px;
  color: #334155;
  cursor: pointer;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
}

.cms-content-model strong {
  min-width: 28px;
  color: #0f172a;
  text-align: right;
}

.cms-content-model.is-active {
  border-color: #14b8a6;
  box-shadow: 0 0 0 2px rgba(20, 184, 166, 0.14);
}

.cms-content-model--blue.is-active {
  background: #eff6ff;
}

.cms-content-model--amber.is-active {
  background: #fffbeb;
}

.cms-content-category-tree {
  padding-top: 10px;
  margin-top: 10px;
  border-top: 1px solid #e2e8f0;
}

.cms-content-category-row {
  display: flex;
  align-items: center;
  gap: 4px;
  min-height: 36px;
  border: 1px solid transparent;
  border-radius: 7px;
}

.cms-content-category-row.is-active {
  background: #f0fdfa;
  border-color: #99f6e4;
}

.cms-content-category-row--child {
  font-size: 13px;
}

.cms-content-category-toggle,
.cms-content-category-spacer {
  flex: 0 0 24px;
  width: 24px;
  height: 24px;
}

.cms-content-category-toggle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: #64748b;
  cursor: pointer;
  background: #f8fafc;
  border: 1px solid #dbe5ef;
  border-radius: 6px;
}

.cms-content-category-toggle:hover {
  color: #0f766e;
  border-color: #99f6e4;
}

.cms-content-category {
  display: flex;
  flex: 1;
  gap: 8px;
  align-items: center;
  justify-content: space-between;
  min-width: 0;
  min-height: 34px;
  padding: 7px 10px;
  color: #475569;
  cursor: pointer;
  background: transparent;
  border: 0;
}

.cms-content-category span:first-child {
  min-width: 0;
  overflow: hidden;
  text-align: left;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.cms-content-category-row.is-active .cms-content-category {
  color: #0f766e;
}

.cms-primary-cell {
  display: flex;
  flex-direction: column;
  gap: 3px;
  min-width: 0;
}

.cms-primary-cell strong {
  overflow: hidden;
  color: #111827;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.cms-primary-cell span {
  overflow: hidden;
  font-size: 12px;
  color: #64748b;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.cms-primary-cell .cms-template-line {
  color: #0f766e;
}

.cms-article-table {
  max-width: 100%;
}

.cms-article-table :deep(.ant-table) {
  table-layout: fixed;
}

.cms-article-table :deep(.ant-table-cell) {
  overflow: hidden;
  vertical-align: middle;
}

.cms-article-table :deep(.ant-table-cell:last-child) {
  padding-right: 10px;
  padding-left: 10px;
}

.cms-article-title-cell,
.cms-article-title-text {
  max-width: 100%;
}

.cms-table-ellipsis {
  display: block;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.cms-article-form {
  padding-bottom: 4px;
}

.cms-message-strip {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
  margin-bottom: 14px;
}

.cms-message-card {
  min-width: 0;
  padding: 12px;
  background: #fffbeb;
  border: 1px solid #fde68a;
  border-radius: 8px;
}

.cms-message-card div,
.cms-message-detail-head {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}

.cms-message-card span,
.cms-message-detail-head span {
  font-size: 12px;
  color: #64748b;
}

.cms-message-card p {
  display: -webkit-box;
  margin: 8px 0 0;
  overflow: hidden;
  color: #334155;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.cms-message-detail {
  padding: 12px;
  margin-bottom: 16px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
}

.cms-message-detail p {
  margin: 10px 0 0;
  color: #334155;
  overflow-wrap: anywhere;
}

:deep(.ant-tabs-nav) {
  display: none;
}

@media (max-width: 1320px) {
  .cms-overview,
  .cms-site-layout {
    grid-template-columns: 1fr;
  }

  .cms-metric-grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }

  .cms-content-layout {
    grid-template-columns: 1fr;
  }

  .cms-content-category-tree {
    max-height: 300px;
    overflow: auto;
  }
}

@media (max-width: 780px) {
  .cms-overview,
  .cms-panel,
  .cms-quickbar {
    padding: 12px;
  }

  .cms-quickbar,
  .cms-panel-head {
    align-items: stretch;
    flex-direction: column;
  }

  .cms-metric-grid,
  .cms-form-grid,
  .cms-site-media-grid,
  .cms-message-strip {
    grid-template-columns: 1fr;
  }

  .cms-filter-input,
  .cms-filter-select {
    width: 100%;
  }

  .cms-article-filterbar {
    grid-template-columns: 1fr;
  }
}
</style>

<style>
.cms-article-modal-wrap .ant-modal {
  top: 48px;
  max-width: calc(100vw - 32px);
  padding-bottom: 48px;
}

.cms-article-modal-wrap .ant-modal-content {
  display: flex;
  max-height: calc(100vh - 96px);
  flex-direction: column;
}

.cms-article-modal-wrap .ant-modal-body {
  flex: 1 1 auto;
  max-height: calc(100vh - 220px);
  padding-right: 18px;
  overflow-y: auto;
}
</style>
