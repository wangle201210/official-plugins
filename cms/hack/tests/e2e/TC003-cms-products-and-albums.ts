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

async function createTypedCategory(
  adminApi: APIRequestContext,
  suffix: number,
  type: number,
  label: string,
): Promise<{ id: number; path: string }> {
  const path = `/e2e-${label}-${suffix}`;
  const created = await expectSuccess<{ id: number }>(
    await adminApi.post("cms/categories", {
      data: {
        code: `e2e-${label}-${suffix}`,
        name: `E2E ${label} 栏目 ${suffix}`,
        path,
        status: 1,
        type,
      },
    }),
  );
  return { id: created.id, path };
}

test.describe("TC-3 CMS 产品中心与相册", () => {
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

  test("TC-3a: 产品管理页签 CRUD 与公开产品流程", async ({ adminPage }) => {
    test.setTimeout(120_000);

    const suffix = Date.now();
    const category = await createTypedCategory(adminApi, suffix, 4, "product");
    const product = {
      name: `E2E 产品 ${suffix}`,
      price: `¥${suffix % 1000}`,
      slug: `e2e-product-${suffix}`,
    };
    let productID = 0;

    try {
      const cmsPage = new CmsPluginPage(adminPage);
      await cmsPage.goto();
      await cmsPage.openProductsTab("产品");
      await cmsPage.createProduct({
        categoryName: `E2E product 栏目 ${suffix}`,
        name: product.name,
        price: product.price,
        slug: product.slug,
        spec: "E2E 规格 1.0",
        summary: `E2E 产品摘要 ${suffix}`,
      });

      const listed = await expectSuccess<{
        list: { id: number; name: string; status: number }[];
        total: number;
      }>(
        await adminApi.get(
          `cms/products?pageNum=1&pageSize=10&name=${encodeURIComponent(product.name)}`,
        ),
      );
      expect(listed.total).toBe(1);
      expect(listed.list[0].status).toBe(1);
      productID = listed.list[0].id;

      const publicDetail = await expectSuccess<{
        name: string;
        price: string;
        slug: string;
        views: number;
      }>(await adminApi.get(`cms/public/products/${product.slug}`));
      expect(publicDetail.name).toBe(product.name);
      expect(publicDetail.price).toBe(product.price);

      const listPage = await adminPage.request.get(
        `/cms-site${category.path}/`,
      );
      expect(listPage.status()).toBe(200);
      const listHTML = await listPage.text();
      expect(listHTML).toContain('data-template="product-list"');
      expect(listHTML).toContain(product.name);
      expect(listHTML).toContain(product.price);

      const detailPage = await adminPage.request.get(
        `/cms-site?product=${product.slug}`,
      );
      expect(detailPage.status()).toBe(200);
      const detailHTML = await detailPage.text();
      expect(detailHTML).toContain('data-template="product-detail"');
      expect(detailHTML).toContain("E2E 规格 1.0");

      const sitemap = await adminPage.request.get("/cms-site/sitemap.xml");
      expect(await sitemap.text()).toContain(`product=${product.slug}`);

      await cmsPage.deleteProduct(product.name);
      const afterDelete = await expectSuccess<{ total: number }>(
        await adminApi.get(
          `cms/products?pageNum=1&pageSize=10&name=${encodeURIComponent(product.name)}`,
        ),
      );
      expect(afterDelete.total).toBe(0);
      productID = 0;
    } finally {
      if (productID > 0) {
        await adminApi.delete(`cms/products/${productID}`);
      }
      await adminApi.delete(`cms/categories/${category.id}`);
    }
  });

  test("TC-3b: 相册整册图片保存与公开相册页渲染", async ({ adminPage }) => {
    test.setTimeout(120_000);

    const suffix = Date.now();
    const category = await createTypedCategory(adminApi, suffix, 5, "album");
    const albumName = `E2E 相册 ${suffix}`;
    let albumID = 0;

    try {
      const created = await expectSuccess<{ id: number }>(
        await adminApi.post("cms/albums", {
          data: {
            categoryId: category.id,
            cover: "https://picsum.photos/seed/e2e-album/400/300",
            description: `E2E 相册描述 ${suffix}`,
            images: [
              {
                sort: 1,
                title: `E2E 图片一 ${suffix}`,
                url: "https://picsum.photos/seed/e2e-photo-1/400/300",
              },
              {
                sort: 2,
                title: `E2E 图片二 ${suffix}`,
                url: "https://picsum.photos/seed/e2e-photo-2/400/300",
              },
            ],
            name: albumName,
            sort: 1,
            status: 1,
          },
        }),
      );
      albumID = created.id;

      const cmsPage = new CmsPluginPage(adminPage);
      await cmsPage.goto();
      await cmsPage.expectAlbumRow(albumName, "2");

      const updatedImages = [
        {
          sort: 1,
          title: `E2E 新图 ${suffix}`,
          url: "https://picsum.photos/seed/e2e-photo-3/400/300",
        },
      ];
      await expectSuccess(
        await adminApi.put(`cms/albums/${albumID}`, {
          data: {
            categoryId: category.id,
            cover: "https://picsum.photos/seed/e2e-album/400/300",
            description: `E2E 相册描述 ${suffix}`,
            images: updatedImages,
            name: albumName,
            sort: 1,
            status: 1,
          },
        }),
      );
      const detail = await expectSuccess<{
        imageCount: number;
        images: { title: string }[];
      }>(await adminApi.get(`cms/albums/${albumID}`));
      expect(detail.imageCount).toBe(1);
      expect(detail.images[0].title).toBe(`E2E 新图 ${suffix}`);

      const listPage = await adminPage.request.get(`/cms-site${category.path}/`);
      expect(listPage.status()).toBe(200);
      const listHTML = await listPage.text();
      expect(listHTML).toContain('data-template="album-list"');
      expect(listHTML).toContain(albumName);

      const detailPage = await adminPage.request.get(
        `/cms-site?album=${albumID}`,
      );
      expect(detailPage.status()).toBe(200);
      const detailHTML = await detailPage.text();
      expect(detailHTML).toContain('data-template="album-detail"');
      expect(detailHTML).toContain(`E2E 新图 ${suffix}`);

      const missingPage = await adminPage.request.get(
        `/cms-site?album=${albumID + 99_999}`,
      );
      expect(missingPage.status()).toBe(404);
    } finally {
      if (albumID > 0) {
        await adminApi.delete(`cms/albums/${albumID}`);
      }
      await adminApi.delete(`cms/categories/${category.id}`);
    }
  });
});
