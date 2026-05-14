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
const staleRuntimeI18nCacheKey = "linapro:i18n:runtime:zh-CN";

type CategoryNode = {
  children?: CategoryNode[];
  code: string;
  id: number;
  name?: string;
};

type SiteDetail = {
  address: string;
  contact: string;
  description: string;
  domain: string;
  email: string;
  icp: string;
  keywords: string;
  logo: string;
  name: string;
  phone: string;
  slogan: string;
  status: number;
  weixin: string;
};

function findCategoryID(categories: CategoryNode[], code: string): number {
  for (const category of categories) {
    if (category.code === code) {
      return category.id;
    }
    const childID = findCategoryID(category.children ?? [], code);
    if (childID > 0) {
      return childID;
    }
  }
  return 0;
}

test.describe("TC-231 CMS 插件管理", () => {
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

  test("TC-231a: 旧版简体中文运行时缓存会被内容指纹 ETag 刷新", async ({
    adminPage,
  }) => {
    await adminPage.evaluate((cacheKey) => {
      localStorage.setItem(
        cacheKey,
        JSON.stringify({
          etag: '"zh-CN-1"',
          messages: {
            plugin: {
              cms: {
                fields: {
                  logo: "Logo",
                },
                metrics: {
                  articles: "内容总数",
                  categories: "栏目总数",
                  pendingMessages: "待处理留言",
                  published: "已发布内容",
                },
                overview: {
                  kicker: "Content operations",
                },
              },
            },
          },
          savedAt: Date.now(),
        }),
      );
    }, staleRuntimeI18nCacheKey);

    const cmsPage = new CmsPluginPage(adminPage);
    await cmsPage.goto();
    await cmsPage.expectWorkbenchVisible();

    await expect
      .poll(async () =>
        adminPage.evaluate((cacheKey) => {
          const cache = JSON.parse(localStorage.getItem(cacheKey) || "{}");
          return {
            articleGlyph: cache.messages?.plugin?.cms?.metrics?.articleGlyph,
            categoryGlyph: cache.messages?.plugin?.cms?.metrics?.categoryGlyph,
            etag: cache.etag,
            kicker: cache.messages?.plugin?.cms?.overview?.kicker,
            logo: cache.messages?.plugin?.cms?.fields?.logo,
          };
        }, staleRuntimeI18nCacheKey),
      )
      .toEqual(
        expect.objectContaining({
          articleGlyph: "文",
          categoryGlyph: "栏",
          kicker: "内容运营",
          logo: "站点标识",
        }),
      );
  });

  test("TC-231b: CMS 页面可打开并完成栏目、内容和留言基础流程", async ({
    adminPage,
  }) => {
    test.setTimeout(120_000);

    const suffix = Date.now();
    const category = {
      code: `e2e-${suffix}`,
      name: `E2E CMS 栏目 ${suffix}`,
      path: `/e2e-${suffix}`,
    };
    const article = {
      categoryName: category.name,
      content: `E2E CMS Content ${suffix}`,
      slug: `e2e-cms-${suffix}`,
      title: `E2E CMS 内容 ${suffix}`,
    };
    const parentCategory = {
      code: `e2e-parent-${suffix}`,
      name: `E2E 父栏目 ${suffix}`,
      path: `/e2e-parent-${suffix}`,
    };
    const childCategory = {
      code: `e2e-child-${suffix}`,
      name: `E2E 子栏目 ${suffix}`,
      path: `/e2e-child-${suffix}`,
    };
    const visitorMessage = `E2E CMS visitor message ${suffix}`;
    const importedArticle = {
      content: `E2E 导入公司简介 ${suffix}`,
      slug: `e2e-imported-company-${suffix}`,
      title: `公司简介 E2E ${suffix}`,
    };

    const cmsPage = new CmsPluginPage(adminPage);
    await cmsPage.goto();
    await cmsPage.expectWorkbenchVisible();
    await cmsPage.expectSiteImageUploadsVisible();
    const originalSite = await expectSuccess<SiteDetail>(
      await adminApi.get("cms/site"),
    );
    await expectSuccess(
      await adminApi.put("cms/site", {
        data: {
          ...originalSite,
          logo: "",
          weixin: "",
        },
      }),
    );
    await adminPage.reload({ waitUntil: "domcontentloaded" });
    await cmsPage.expectReady();
    await cmsPage.updateSiteImages();
    const updatedSite = await expectSuccess<SiteDetail>(
      await adminApi.get("cms/site"),
    );
    expect(updatedSite.logo).toContain("/api/v1/uploads/");
    expect(updatedSite.weixin).toContain("/api/v1/uploads/");
    await cmsPage.expectSiteImagesPersisted();
    await cmsPage.expectArticleEditorVisible();
    await cmsPage.expectContentModulesVisible();
    await cmsPage.expectArticleListLayoutStable();
    const slide = await expectSuccess<{ id: number }>(
      await adminApi.post("cms/slides", {
        data: {
          groupCode: "e2e",
          image: "https://example.com/e2e-slide.png",
          sort: 1,
          status: 1,
          subtitle: `E2E 轮播副标题 ${suffix}`,
          title: `E2E 轮播 ${suffix}`,
        },
      }),
    );
    const link = await expectSuccess<{ id: number }>(
      await adminApi.post("cms/links", {
        data: {
          groupCode: "e2e",
          name: `E2E 友情链接 ${suffix}`,
          sort: 1,
          status: 1,
          url: "https://example.com",
        },
      }),
    );
    await cmsPage.expectSlideAndLinkManagersVisible({
      linkName: `E2E 友情链接 ${suffix}`,
      slideTitle: `E2E 轮播 ${suffix}`,
    });
    await cmsPage.createCategory(category);
    await cmsPage.createArticle(article);
    await cmsPage.searchArticle(article.title);
    const createdParent = await expectSuccess<{ id: number }>(
      await adminApi.post("cms/categories", {
        data: {
          code: parentCategory.code,
          name: parentCategory.name,
          path: parentCategory.path,
          status: 1,
          type: 1,
        },
      }),
    );
    const createdChild = await expectSuccess<{ id: number }>(
      await adminApi.post("cms/categories", {
        data: {
          code: childCategory.code,
          name: childCategory.name,
          parentId: createdParent.id,
          path: childCategory.path,
          status: 1,
          type: 1,
        },
      }),
    );
    await cmsPage.expectArticleCategoryTreeCollapsible({
      childId: createdChild.id,
      childName: childCategory.name,
      parentId: createdParent.id,
      parentName: parentCategory.name,
    });
    const categories = await expectSuccess<{ list: CategoryNode[] }>(
      await adminApi.get("cms/categories"),
    );
    const articleCategoryID = findCategoryID(categories.list, category.code);
    const newsCenterID = findCategoryID(categories.list, "5");
    expect(articleCategoryID).toBeGreaterThan(0);
    expect(newsCenterID).toBeGreaterThan(0);
    await cmsPage.expectModelContainsArticle({
      model: "list",
      title: "研究院2025年度工作总结暨表彰大会",
    });
    await cmsPage.expectCategoryGroupContainsArticle({
      categoryId: newsCenterID,
      title: "研究院2025年度工作总结暨表彰大会",
    });

    const publicList = await expectSuccess<{
      list: { slug: string }[];
      total: number;
    }>(
      await adminApi.get(
        `cms/public/articles?pageNum=1&pageSize=20&keyword=${encodeURIComponent(
          article.title,
        )}`,
      ),
    );
    expect(
      publicList.list.some((item) => item.slug === article.slug),
    ).toBeTruthy();

    const publicDetail = await expectSuccess<{
      content: string;
      cover: string;
      slug: string;
    }>(await adminApi.get(`cms/public/articles/${article.slug}`));
    expect(publicDetail.slug).toBe(article.slug);
    expect(publicDetail.content).toContain(article.content);
    expect(publicDetail.cover).toContain("/api/v1/uploads/");

    const imported = await expectSuccess<{ id: number }>(
      await adminApi.post("cms/articles", {
        data: {
          categoryId: articleCategoryID,
          content: `&lt;p&gt;&lt;span style=&quot;font-family: SimSun; font-size: 18px;&quot;&gt;${importedArticle.content}&lt;/span&gt;&lt;/p&gt;`,
          slug: importedArticle.slug,
          status: 1,
          title: importedArticle.title,
        },
      }),
    );
    const importedDetail = await expectSuccess<{ content: string }>(
      await adminApi.get(`cms/articles/${imported.id}`),
    );
    expect(importedDetail.content).toContain("<p>");
    expect(importedDetail.content).toContain(importedArticle.content);
    expect(importedDetail.content).not.toContain("&lt;p");
    await cmsPage.expectArticleSourceEditor(importedArticle);

    await expectSuccess<{ id: number }>(
      await adminApi.post("cms/public/messages", {
        data: {
          content: visitorMessage,
          email: `cms-${suffix}@example.com`,
          name: `CMS Visitor ${suffix}`,
        },
      }),
    );
    await cmsPage.approveMessage(visitorMessage);

    await cmsPage.deleteArticle(importedArticle.title);
    await cmsPage.deleteArticle(article.title);
    await cmsPage.deleteCategory(category.name);
    await expectSuccess<unknown>(
      await adminApi.delete(`cms/categories/${createdChild.id}`),
    );
    await expectSuccess<unknown>(
      await adminApi.delete(`cms/categories/${createdParent.id}`),
    );
    await expectSuccess<unknown>(await adminApi.delete(`cms/links/${link.id}`));
    await expectSuccess<unknown>(
      await adminApi.delete(`cms/slides/${slide.id}`),
    );
    await expectSuccess<unknown>(
      await adminApi.put("cms/site", { data: originalSite }),
    );
  });
});
