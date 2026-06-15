import type { APIRequestContext } from "../../../../../../hack/tests/support/playwright";

import { test, expect } from "../../../../../../hack/tests/fixtures/auth";
import { refreshPluginProjection } from "../../../../../../hack/tests/fixtures/plugin";
import { CmsPluginPage } from "../pages/CmsPluginPage";
import {
  createAdminApiContext,
  enablePlugin,
  expectSuccess,
  getPlugin,
  installPlugin,
  syncPlugins,
} from "../../../../../../hack/tests/support/api/job";

const pluginID = "cms";

type ArticleDetail = {
  id: number;
  publishedAt?: number;
  slug: string;
  status: number;
};

async function createCategory(
  adminApi: APIRequestContext,
  suffix: number,
): Promise<number> {
  const created = await expectSuccess<{ id: number }>(
    await adminApi.post("cms/categories", {
      data: {
        code: `e2e-batch-${suffix}`,
        name: `E2E 批量栏目 ${suffix}`,
        path: `/e2e-batch-${suffix}`,
        status: 1,
        type: 1,
      },
    }),
  );
  return created.id;
}

async function createDraftArticle(
  adminApi: APIRequestContext,
  categoryID: number,
  slug: string,
  title: string,
): Promise<number> {
  const created = await expectSuccess<{ id: number }>(
    await adminApi.post("cms/articles", {
      data: {
        categoryId: categoryID,
        content: `<p>${title}</p>`,
        slug,
        status: 0,
        title,
      },
    }),
  );
  return created.id;
}

async function getArticle(
  adminApi: APIRequestContext,
  id: number,
): Promise<ArticleDetail> {
  return expectSuccess<ArticleDetail>(await adminApi.get(`cms/articles/${id}`));
}

test.describe("TC-2 CMS 文章批量操作、定时发布与公开 SEO 端点", () => {
  let adminApi: APIRequestContext;

  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
    await syncPlugins(adminApi);
  });

  test.afterAll(async () => {
    await adminApi.dispose();
  });

  test.beforeEach(async ({ adminPage }) => {
    await syncPlugins(adminApi);
    let plugin = await getPlugin(adminApi, pluginID);
    if (plugin.installed !== 1) {
      await installPlugin(adminApi, pluginID);
      plugin = await getPlugin(adminApi, pluginID);
    }
    if (plugin.enabled !== 1) {
      await enablePlugin(adminApi, pluginID);
    }
    await refreshPluginProjection(adminPage);
  });

  test("TC-2a: 文章列表批量发布、批量下线和批量删除闭环", async ({
    adminPage,
  }) => {
    test.setTimeout(120_000);

    const suffix = Date.now();
    const categoryID = await createCategory(adminApi, suffix);
    const titles = [1, 2].map((index) => `E2E 批量文章 ${suffix}-${index}`);
    const ids: number[] = [];
    for (const [index, title] of titles.entries()) {
      ids.push(
        await createDraftArticle(
          adminApi,
          categoryID,
          `e2e-batch-${suffix}-${index + 1}`,
          title,
        ),
      );
    }

    try {
      const cmsPage = new CmsPluginPage(adminPage);
      await cmsPage.goto();
      await cmsPage.expectBatchButtonsDisabledWithoutSelection({
        batchDelete: "批量删除",
        batchPublish: "批量发布",
        batchUnpublish: "批量下线",
      });

      await cmsPage.filterArticlesByTitle(`E2E 批量文章 ${suffix}`);
      for (const title of titles) {
        await cmsPage.selectArticleRow(title);
      }
      await cmsPage.expectSelectedArticlesCount("已选 2 项");
      await cmsPage.batchPublishSelectedArticles();
      for (const id of ids) {
        const published = await getArticle(adminApi, id);
        expect(published.status).toBe(1);
        expect(published.publishedAt).toBeTruthy();
      }

      await cmsPage.filterArticlesByTitle(`E2E 批量文章 ${suffix}`);
      for (const title of titles) {
        await cmsPage.selectArticleRow(title);
      }
      await cmsPage.expectSelectedArticlesCount("已选 2 项");
      await cmsPage.batchUnpublishSelectedArticles();
      for (const id of ids) {
        const unpublished = await getArticle(adminApi, id);
        expect(unpublished.status).toBe(0);
        expect(unpublished.publishedAt).toBeTruthy();
      }

      await cmsPage.filterArticlesByTitle(`E2E 批量文章 ${suffix}`);
      for (const title of titles) {
        await cmsPage.selectArticleRow(title);
      }
      await cmsPage.expectSelectedArticlesCount("已选 2 项");
      await cmsPage.batchDeleteSelectedArticles();
      const remaining = await expectSuccess<{ total: number }>(
        await adminApi.get(
          `cms/articles?pageNum=1&pageSize=10&title=${encodeURIComponent(
            `E2E 批量文章 ${suffix}`,
          )}`,
        ),
      );
      expect(remaining.total).toBe(0);
      ids.length = 0;
    } finally {
      for (const id of ids) {
        await adminApi.delete(`cms/articles/${id}`);
      }
      await adminApi.delete(`cms/categories/${categoryID}`);
    }
  });

  test("TC-2b: 定时发布文章在到点前对公开站点不可见", async ({ adminPage }) => {
    test.setTimeout(120_000);

    const suffix = Date.now();
    const categoryID = await createCategory(adminApi, suffix);
    const scheduled = {
      content: `E2E 定时内容 ${suffix}`,
      slug: `e2e-scheduled-${suffix}`,
      title: `E2E 定时文章 ${suffix}`,
    };
    const released = {
      slug: `e2e-released-${suffix}`,
      title: `E2E 已到点文章 ${suffix}`,
    };

    let releasedID = 0;
    try {
      const cmsPage = new CmsPluginPage(adminPage);
      await cmsPage.goto();
      await cmsPage.createScheduledArticle({
        categoryName: `E2E 批量栏目 ${suffix}`,
        content: scheduled.content,
        publishedAtLocal: "2030-01-01T08:00",
        slug: scheduled.slug,
        title: scheduled.title,
      });
      await cmsPage.expectArticleScheduledTag(scheduled.title, "定时");

      const hiddenDetail = await adminApi.get(
        `cms/public/articles/${scheduled.slug}`,
      );
      const hiddenPayload = (await hiddenDetail.json()) as { code: number };
      expect(hiddenPayload.code).not.toBe(0);
      const publicList = await expectSuccess<{
        list: { slug: string }[];
      }>(
        await adminApi.get(
          `cms/public/articles?pageNum=1&pageSize=20&keyword=${encodeURIComponent(
            scheduled.title,
          )}`,
        ),
      );
      expect(
        publicList.list.some((item) => item.slug === scheduled.slug),
      ).toBeFalsy();

      const releasedCreated = await expectSuccess<{ id: number }>(
        await adminApi.post("cms/articles", {
          data: {
            categoryId: categoryID,
            content: `<p>${released.title}</p>`,
            publishedAt: Date.now() - 3_600_000,
            slug: released.slug,
            status: 1,
            title: released.title,
          },
        }),
      );
      releasedID = releasedCreated.id;
      const releasedDetail = await expectSuccess<{ slug: string }>(
        await adminApi.get(`cms/public/articles/${released.slug}`),
      );
      expect(releasedDetail.slug).toBe(released.slug);
    } finally {
      const list = await expectSuccess<{ list: { id: number }[] }>(
        await adminApi.get(
          `cms/articles?pageNum=1&pageSize=20&title=${encodeURIComponent(
            scheduled.title,
          )}`,
        ),
      );
      for (const item of list.list) {
        await adminApi.delete(`cms/articles/${item.id}`);
      }
      if (releasedID > 0) {
        await adminApi.delete(`cms/articles/${releasedID}`);
      }
      await adminApi.delete(`cms/categories/${categoryID}`);
    }
  });

  test("TC-2c: sitemap.xml、rss.xml 和 robots.txt 输出公开内容", async ({
    adminPage,
  }) => {
    test.setTimeout(120_000);

    const suffix = Date.now();
    const categoryID = await createCategory(adminApi, suffix);
    const visible = {
      slug: `e2e-seo-visible-${suffix}`,
      title: `E2E SEO 可见文章 ${suffix}`,
    };
    const scheduledSlug = `e2e-seo-scheduled-${suffix}`;
    let visibleID = 0;
    let scheduledID = 0;

    try {
      const visibleCreated = await expectSuccess<{ id: number }>(
        await adminApi.post("cms/articles", {
          data: {
            categoryId: categoryID,
            content: `<p>${visible.title}</p>`,
            publishedAt: Date.now() - 3_600_000,
            slug: visible.slug,
            status: 1,
            summary: `E2E SEO 摘要 ${suffix}`,
            title: visible.title,
          },
        }),
      );
      visibleID = visibleCreated.id;
      const scheduledCreated = await expectSuccess<{ id: number }>(
        await adminApi.post("cms/articles", {
          data: {
            categoryId: categoryID,
            content: `<p>scheduled</p>`,
            publishedAt: Date.now() + 86_400_000,
            slug: scheduledSlug,
            status: 1,
            title: `E2E SEO 定时文章 ${suffix}`,
          },
        }),
      );
      scheduledID = scheduledCreated.id;

      const sitemap = await adminPage.request.get("/cms-site/sitemap.xml");
      expect(sitemap.status()).toBe(200);
      expect(sitemap.headers()["content-type"]).toContain("application/xml");
      const sitemapText = await sitemap.text();
      expect(sitemapText).toContain("<urlset");
      expect(sitemapText).toContain(`article=${visible.slug}`);
      expect(sitemapText).toContain(`/e2e-batch-${suffix}/`);
      expect(sitemapText).not.toContain(scheduledSlug);

      const rss = await adminPage.request.get("/cms-site/rss.xml");
      expect(rss.status()).toBe(200);
      expect(rss.headers()["content-type"]).toContain("application/xml");
      const rssText = await rss.text();
      expect(rssText).toContain('<rss version="2.0"');
      expect(rssText).toContain(visible.title);
      expect(rssText).toContain(`E2E SEO 摘要 ${suffix}`);
      expect(rssText).not.toContain(scheduledSlug);

      const robots = await adminPage.request.get("/cms-site/robots.txt");
      expect(robots.status()).toBe(200);
      expect(robots.headers()["content-type"]).toContain("text/plain");
      const robotsText = await robots.text();
      expect(robotsText).toContain("User-agent: *");
      expect(robotsText).toContain("Sitemap:");
      expect(robotsText).toContain("/cms-site/sitemap.xml");
    } finally {
      if (visibleID > 0) {
        await adminApi.delete(`cms/articles/${visibleID}`);
      }
      if (scheduledID > 0) {
        await adminApi.delete(`cms/articles/${scheduledID}`);
      }
      await adminApi.delete(`cms/categories/${categoryID}`);
    }
  });
});
