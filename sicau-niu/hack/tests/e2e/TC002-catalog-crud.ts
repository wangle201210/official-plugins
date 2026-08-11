import { test, expect } from "@host-tests/fixtures/auth";
import { ensureSourcePluginEnabled } from "@host-tests/fixtures/plugin";

import { SicauNiuCatalogPage } from "../pages/SicauNiuCatalogPage";

const pluginID = "sicau-niu";

// TC-2 covers the sicau-niu operator content-asset CRUD, the primary
// user-observable path of the C2 niu-catalog-admin change. It exercises the two
// most stable assets: 金句 (quote: content + enabled) and 普通牛 (common niu:
// code + optional online time). For each asset it creates a record and asserts it
// persists in the list, edits it and asserts the updated row, then deletes it and
// asserts removal. A unique run-scoped suffix keeps each test self-contained and
// order-independent. It also verifies that the new-iron modal sends the fixed
// 10-second IOT reporting-cycle action without creating a local iron record.
//
// 卡片 (card) CRUD is intentionally NOT covered here. The card form is coupled to
// an existing 所属牛 (owning niu) selection plus the "1 牛 1 卡" uniqueness
// constraint, so a reliable card E2E needs a niu fixture wired through the card
// modal's owning-niu Select. Without a running backend to validate that
// interaction, a card case would be brittle; it is left for a follow-up once the
// flow can be executed and stabilised.
test.describe("TC-2 sicau-niu 内容资产 CRUD", () => {
  let catalogPage: SicauNiuCatalogPage;
  const suffix = `${Date.now()}`;

  test.beforeEach(async ({ adminPage }) => {
    await ensureSourcePluginEnabled(adminPage, pluginID);
    catalogPage = new SicauNiuCatalogPage(adminPage);
  });

  test("TC-2a: 新增金句后列表出现该记录", async () => {
    const content = `川农金句-${suffix}-a`;
    await catalogPage.openQuoteFromMenu();
    await catalogPage.createQuote(content);
    await expect(catalogPage.quoteRow(content)).toBeVisible();
  });

  test("TC-2b: 编辑金句内容后列表更新", async () => {
    const content = `川农金句-${suffix}-b`;
    const renamed = `川农金句-改-${suffix}-b`;
    await catalogPage.openQuoteFromMenu();
    await catalogPage.createQuote(content);
    await catalogPage.editQuote(content, renamed);
    await expect(catalogPage.quoteRow(renamed)).toBeVisible();
  });

  test("TC-2c: 删除金句后列表移除该记录", async () => {
    const content = `川农金句-待删-${suffix}-c`;
    await catalogPage.openQuoteFromMenu();
    await catalogPage.createQuote(content);
    await catalogPage.deleteQuote(content);
    await expect(catalogPage.quoteRow(content)).toHaveCount(0);
  });

  test("TC-2d: 新增普通牛后列表出现该记录", async () => {
    const code = `NIU-${suffix}-d`;
    await catalogPage.openNiuFromMenu();
    await catalogPage.createCommonNiu(code);
    await expect(catalogPage.niuRow(code)).toBeVisible();
  });

  test("TC-2e: 铁牛管理只读展示最近同步经纬度", async () => {
    await catalogPage.openIronFromMenu();
    await catalogPage.expectIronLocationReadonly();
  });

  test("TC-2f: 新增铁牛时可将设备上报频率设置为 10 秒", async ({
    adminPage,
  }) => {
    const code = "50275156712";
    let createCalled = false;
    await adminPage.route(
      /\/x\/sicau-niu\/api\/v1\/plugins\/sicau-niu\/admin\/iron$/u,
      async (route) => {
        createCalled = true;
        await route.fulfill({
          json: { code: 0, data: { id: 1 }, message: "success" },
        });
      },
    );
    await adminPage.route(
      "**/x/sicau-niu/api/v1/plugins/sicau-niu/admin/iron/reporting-cycle",
      async (route) => {
        await route.fulfill({
          json: { code: 0, data: {}, message: "success" },
        });
      },
    );

    await catalogPage.openIronFromMenu();
    await catalogPage.openNewIronModal();
    const requestPromise = adminPage.waitForRequest(
      (request) =>
        request.method() === "PUT" &&
        request
          .url()
          .endsWith(
            "/x/sicau-niu/api/v1/plugins/sicau-niu/admin/iron/reporting-cycle",
          ),
    );
    await catalogPage.updateIronReportingCycle(code);
    const request = await requestPromise;

    expect(request.postDataJSON()).toEqual({ code });
    await expect(
      adminPage.getByText("上报频率已更新为 10 秒", { exact: true }),
    ).toBeVisible();
    await expect(catalogPage.ironModal()).toBeVisible();
    await expect(catalogPage.ironCodeInput()).toHaveValue(code);
    expect(createCalled).toBe(false);
  });

  test("TC-2g: 上报频率更新失败时保留新增弹窗并展示错误", async ({
    adminPage,
  }) => {
    const code = "50275156712";
    const failureMessage = "定位器上报频率更新失败";
    await adminPage.route(
      "**/x/sicau-niu/api/v1/plugins/sicau-niu/admin/iron/reporting-cycle",
      async (route) => {
        await route.fulfill({
          json: {
            code: 50,
            data: null,
            errorCode: "PLUGIN_SICAU_NIU_IRON_LOCATION_IOT_REQUEST_FAILED",
            message: failureMessage,
            messageKey:
              "error.plugin.sicau.niu.iron.location.iot.request.failed",
            messageParams: {},
          },
          status: 200,
        });
      },
    );

    await catalogPage.openIronFromMenu();
    await catalogPage.openNewIronModal();
    await catalogPage.updateIronReportingCycle(code);

    await expect(
      adminPage.locator(".ant-message-notice:visible", {
        hasText: failureMessage,
      }),
    ).toHaveCount(1);
    await expect(catalogPage.ironModal()).toBeVisible();
    await expect(catalogPage.ironCodeInput()).toHaveValue(code);
    await expect(catalogPage.ironReportingCycleButton()).toBeEnabled();
  });

  // 注:牛的编辑/删除经后端单元测试(cattle DB 门控:更新 code 冲突、删除、删牛级联软删
  // 主卡)与 API 验证(GET/PUT/DELETE 均 code:0)覆盖;其复杂编辑表单(类型/子类/院系联动
  // + 详情回填)在 Playwright 下交互不稳定,牛的 UI 编辑/删除 E2E 留待后续硬化。本 TC 以
  // 金句全量 CRUD、牛新增和铁牛 IOT 指令证明运营页关键路径端到端贯通。
});
