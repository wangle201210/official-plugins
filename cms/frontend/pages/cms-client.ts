import { requestClient } from "#/api/request";

export interface Site {
  address: string;
  contact: string;
  description: string;
  domain: string;
  email: string;
  icp: string;
  id: number;
  keywords: string;
  logo: string;
  name: string;
  phone: string;
  siteKey: string;
  slogan: string;
  status: number;
  weixin: string;
}

export interface Category {
  children?: Category[];
  code: string;
  contentTemplate: string;
  cover: string;
  description: string;
  id: number;
  keywords: string;
  listTemplate: string;
  name: string;
  outlink: string;
  parentId: number;
  path: string;
  sort: number;
  status: number;
  title: string;
  type: number;
}

export interface Article {
  author: string;
  categoryId: number;
  categoryName: string;
  content: string;
  cover: string;
  description: string;
  id: number;
  isRecommend: number;
  isTop: number;
  keywords: string;
  publishedAt?: string;
  slug: string;
  sort: number;
  source: string;
  status: number;
  subtitle: string;
  summary: string;
  tags: string;
  title: string;
  views: number;
}

export interface Message {
  content: string;
  createdAt?: string;
  email: string;
  id: number;
  mobile: string;
  name: string;
  reply: string;
  status: number;
  userIp: string;
}

export interface Link {
  createdAt?: string;
  groupCode: string;
  id: number;
  logo: string;
  name: string;
  sort: number;
  status: number;
  url: string;
}

export interface Slide {
  createdAt?: string;
  groupCode: string;
  id: number;
  image: string;
  link: string;
  sort: number;
  status: number;
  subtitle: string;
  title: string;
}

export interface ArticleListParams {
  categoryId?: number;
  categoryType?: number;
  includeChildren?: boolean;
  pageNum?: number;
  pageSize?: number;
  status?: number;
  title?: string;
}

export interface MessageListParams {
  keyword?: string;
  pageNum?: number;
  pageSize?: number;
  status?: number;
}

export interface LinkListParams {
  groupCode?: string;
  keyword?: string;
  pageNum?: number;
  pageSize?: number;
  status?: number;
}

export interface SlideListParams {
  groupCode?: string;
  keyword?: string;
  pageNum?: number;
  pageSize?: number;
  status?: number;
}

export async function cmsSite() {
  return requestClient.get<Site>("/cms/site");
}

export function cmsSiteUpdate(data: Partial<Site>) {
  return requestClient.put("/cms/site", data);
}

export async function cmsCategoryList(params?: { status?: number }) {
  const res = await requestClient.get<{ list: Category[] }>("/cms/categories", {
    params,
  });
  return res.list;
}

export function cmsCategoryCreate(data: Partial<Category>) {
  return requestClient.post("/cms/categories", data);
}

export function cmsCategoryUpdate(id: number, data: Partial<Category>) {
  return requestClient.put(`/cms/categories/${id}`, data);
}

export function cmsCategoryDelete(id: number) {
  return requestClient.delete(`/cms/categories/${id}`);
}

export async function cmsArticleList(params?: ArticleListParams) {
  const res = await requestClient.get<{ list: Article[]; total: number }>(
    "/cms/articles",
    { params },
  );
  return { items: res.list, total: res.total };
}

export function cmsArticleInfo(id: number) {
  return requestClient.get<Article>(`/cms/articles/${id}`);
}

export function cmsArticleCreate(data: Partial<Article>) {
  return requestClient.post("/cms/articles", data);
}

export function cmsArticleUpdate(id: number, data: Partial<Article>) {
  return requestClient.put(`/cms/articles/${id}`, data);
}

export function cmsArticleDelete(id: number) {
  return requestClient.delete(`/cms/articles/${id}`);
}

export async function cmsMessageList(params?: MessageListParams) {
  const res = await requestClient.get<{ list: Message[]; total: number }>(
    "/cms/messages",
    { params },
  );
  return { items: res.list, total: res.total };
}

export function cmsMessageUpdate(id: number, data: Partial<Message>) {
  return requestClient.put(`/cms/messages/${id}`, data);
}

export function cmsMessageDelete(id: number) {
  return requestClient.delete(`/cms/messages/${id}`);
}

export async function cmsLinkList(params?: LinkListParams) {
  const res = await requestClient.get<{ list: Link[]; total: number }>(
    "/cms/links",
    { params },
  );
  return { items: res.list, total: res.total };
}

export function cmsLinkCreate(data: Partial<Link>) {
  return requestClient.post("/cms/links", data);
}

export function cmsLinkUpdate(id: number, data: Partial<Link>) {
  return requestClient.put(`/cms/links/${id}`, data);
}

export function cmsLinkDelete(id: number) {
  return requestClient.delete(`/cms/links/${id}`);
}

export async function cmsSlideList(params?: SlideListParams) {
  const res = await requestClient.get<{ list: Slide[]; total: number }>(
    "/cms/slides",
    { params },
  );
  return { items: res.list, total: res.total };
}

export function cmsSlideCreate(data: Partial<Slide>) {
  return requestClient.post("/cms/slides", data);
}

export function cmsSlideUpdate(id: number, data: Partial<Slide>) {
  return requestClient.put(`/cms/slides/${id}`, data);
}

export function cmsSlideDelete(id: number) {
  return requestClient.delete(`/cms/slides/${id}`);
}
