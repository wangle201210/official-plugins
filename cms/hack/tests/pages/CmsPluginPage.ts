import { Buffer } from "node:buffer";

import type { Locator, Page } from "../../../../../../hack/tests/support/playwright";

import { expect } from "../../../../../../hack/tests/support/playwright";

import {
  waitForBusyIndicatorsToClear,
  waitForConfirmOverlay,
  waitForDialogReady,
  waitForRouteReady,
  waitForUploadReady,
} from "../../../../../../hack/tests/support/ui";

const coverPngBuffer = Buffer.from(
  "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/p94AAAAASUVORK5CYII=",
  "base64",
);

export class CmsPluginPage {
  constructor(private page: Page) {}

  private get categoryModal(): Locator {
    return this.page
      .locator(".ant-modal:visible", { hasText: /栏目|Category/i })
      .last();
  }

  private get articleModal(): Locator {
    return this.page
      .locator(".ant-modal:visible", { hasText: /内容|Article/i })
      .last();
  }

  private get messageModal(): Locator {
    return this.page
      .locator(".ant-modal:visible", { hasText: /处理留言|Review Message/i })
      .last();
  }

  async goto() {
    await this.page.goto("/cms", { waitUntil: "domcontentloaded" });
    await this.expectReady();
  }

  async expectReady() {
    await this.page.getByTestId("cms-plugin-tabs").waitFor({
      state: "visible",
      timeout: 15_000,
    });
    await this.page.getByTestId("cms-workbench").waitFor({
      state: "visible",
      timeout: 15_000,
    });
    await waitForRouteReady(this.page, 15_000);
  }

  async expectWorkbenchVisible() {
    await this.page.getByTestId("cms-overview").waitFor({
      state: "visible",
      timeout: 10_000,
    });
    await this.page.getByText("内容运营", { exact: true }).waitFor({
      state: "visible",
      timeout: 10_000,
    });
    await this.page.getByText("站点标识", { exact: true }).waitFor({
      state: "visible",
      timeout: 10_000,
    });
    await expect(
      this.page.getByText("Content operations", { exact: true }),
    ).toHaveCount(0);
    await expect(this.page.getByText("Logo", { exact: true })).toHaveCount(0);
    await this.page.getByTestId("cms-metric-categories").waitFor({
      state: "visible",
      timeout: 10_000,
    });
    await this.page.getByTestId("cms-metric-articles").waitFor({
      state: "visible",
      timeout: 10_000,
    });
    await this.page.getByTestId("cms-metric-pending-messages").waitFor({
      state: "visible",
      timeout: 10_000,
    });
  }

  async expectSiteImageUploadsVisible() {
    await this.page.getByTestId("cms-section-site").click();
    await this.page.getByTestId("cms-site-show-messages").waitFor({
      state: "visible",
      timeout: 10_000,
    });
    await expect(this.page.getByTestId("cms-site-show-messages")).toHaveClass(
      /ant-switch/,
    );
    await expect(this.page.getByTestId("cms-site-show-messages")).toHaveAttribute(
      "role",
      "switch",
    );
    await expect(this.page.getByTestId("cms-site-clear-data")).toBeVisible();
    await expect(this.page.getByTestId("cms-site-clear-data")).toContainText(
      /清空数据|Clear Data|清空資料/,
    );
    await expect(this.page.getByTestId("cms-site-clear-data")).not.toContainText(
      "plugin.cms.actions.clearData",
    );
    await expect(this.page.getByTestId("cms-site-load-sample-data")).toBeVisible();
    await expect(
      this.page.getByTestId("cms-site-load-sample-data"),
    ).toContainText(/加载示例数据|Load Sample Data|載入示例資料/);
    await expect(
      this.page.getByTestId("cms-site-load-sample-data"),
    ).not.toContainText("plugin.cms.actions.loadSampleData");
    await this.page.getByTestId("cms-site-logo-upload").waitFor({
      state: "visible",
      timeout: 10_000,
    });
    await this.page.getByTestId("cms-site-weixin-upload").waitFor({
      state: "visible",
      timeout: 10_000,
    });
    await expect(
      this.page
        .getByTestId("cms-site-logo-upload")
        .locator('input[type="file"]'),
    ).toBeAttached();
    await expect(
      this.page
        .getByTestId("cms-site-weixin-upload")
        .locator('input[type="file"]'),
    ).toBeAttached();
    await expect
      .poll(async () =>
        this.page.evaluate(() => {
          const logo = document
            .querySelector('[data-testid="cms-site-logo-upload"]')
            ?.getBoundingClientRect();
          const weixin = document
            .querySelector('[data-testid="cms-site-weixin-upload"]')
            ?.getBoundingClientRect();
          return {
            inline:
              !!logo &&
              !!weixin &&
              Math.abs(logo.top - weixin.top) <= 6 &&
              logo.left < weixin.left,
          };
        }),
      )
      .toEqual({ inline: true });
  }

  async updateSiteImages() {
    await this.page.getByTestId("cms-section-site").click();
    await this.uploadSiteImage("cms-site-logo-upload", "cms-site-logo.png");
    await this.uploadSiteImage("cms-site-weixin-upload", "cms-site-weixin.png");
    const saveResponse = this.page.waitForResponse(
      (response) =>
        response.url().includes("/api/v1/cms/site") &&
        response.request().method() === "PUT" &&
        response.status() === 200,
    );
    await this.page.getByTestId("cms-site-save").click();
    const response = await saveResponse;
    const payload = response.request().postDataJSON() as {
      logo?: string;
      weixin?: string;
    };
    expect(payload.logo).toContain("/api/v1/uploads/");
    expect(payload.weixin).toContain("/api/v1/uploads/");
    await waitForRouteReady(this.page);
  }

  async expectSiteImagesPersisted() {
    await this.page.reload({ waitUntil: "domcontentloaded" });
    await this.expectReady();
    await this.page.getByTestId("cms-section-site").click();
    await expect(
      this.page.getByTestId("cms-site-logo-upload").locator(".ant-upload-list-item"),
    ).toHaveCount(1);
    await expect(
      this.page
        .getByTestId("cms-site-weixin-upload")
        .locator(".ant-upload-list-item"),
    ).toHaveCount(1);
  }

  async openCategoriesTab() {
    await this.page.getByTestId("cms-section-categories").click();
    await this.page.getByTestId("cms-category-table").waitFor({
      state: "visible",
      timeout: 10_000,
    });
    await waitForBusyIndicatorsToClear(this.page);
  }

  async openArticlesTab() {
    await this.page.getByTestId("cms-section-articles").click();
    await this.page.getByTestId("cms-article-table").waitFor({
      state: "visible",
      timeout: 10_000,
    });
    await waitForBusyIndicatorsToClear(this.page);
  }

  async openMessagesTab() {
    await this.page.getByTestId("cms-section-messages").click();
    await this.page.getByTestId("cms-message-table").waitFor({
      state: "visible",
      timeout: 10_000,
    });
    await waitForBusyIndicatorsToClear(this.page);
  }

  async openSlidesTab() {
    await this.page.getByTestId("cms-section-slides").click();
    await this.page.getByTestId("cms-slide-table").waitFor({
      state: "visible",
      timeout: 10_000,
    });
    await waitForBusyIndicatorsToClear(this.page);
  }

  async openLinksTab() {
    await this.page.getByTestId("cms-section-links").click();
    await this.page.getByTestId("cms-link-table").waitFor({
      state: "visible",
      timeout: 10_000,
    });
    await waitForBusyIndicatorsToClear(this.page);
  }

  async expectSlideAndLinkManagersVisible(input: {
    linkName: string;
    slideTitle: string;
  }) {
    await this.openSlidesTab();
    await expect(this.page.getByTestId("cms-section-slides")).toContainText(
      /轮播图|Slides/i,
    );
    await expect(this.page.getByTestId("cms-slide-table")).toContainText(
      input.slideTitle,
    );
    await expect(this.page.getByTestId("cms-slide-add-secondary")).toBeVisible();

    await this.openLinksTab();
    await expect(this.page.getByTestId("cms-section-links")).toContainText(
      /友情链接|Links/i,
    );
    await expect(this.page.getByTestId("cms-link-table")).toContainText(
      input.linkName,
    );
    await expect(this.page.getByTestId("cms-link-add-secondary")).toBeVisible();
  }

  async createCategory(input: { code: string; name: string; path: string }) {
    await this.openCategoriesTab();
    await this.page.getByTestId("cms-category-add").click();
    const modal = await waitForDialogReady(this.categoryModal);
    await modal.getByTestId("cms-category-name-input").fill(input.name);
    await modal.getByTestId("cms-category-code-input").fill(input.code);
    await modal.getByTestId("cms-category-path-input").fill(input.path);
    await expect(modal.getByTestId("cms-category-list-template-input")).toBeVisible();
    await expect(
      modal.getByTestId("cms-category-content-template-input"),
    ).toBeVisible();
    await this.selectOptionByText(
      modal.getByTestId("cms-category-list-template-input"),
      /标准列表|Standard List/i,
    );
    await this.selectOptionByText(
      modal.getByTestId("cms-category-content-template-input"),
      /文章详情|Article Detail/i,
    );
    await this.confirmDialog(modal);
    await this.expectTableText("cms-category-table", input.name);
    await this.expectTableText("cms-category-table", "list.html / detail.html");
  }

  async deleteCategory(name: string) {
    await this.openCategoriesTab();
    const row = this.tableRowByText("cms-category-table", name);
    await row.waitFor({ state: "visible", timeout: 10_000 });
    await row.getByRole("button", { name: /删除|Delete/i }).click();
    await this.confirmPopconfirm();
    await waitForRouteReady(this.page);
  }

  async createArticle(input: {
    categoryName: string;
    content: string;
    slug: string;
    title: string;
  }) {
    await this.openArticlesTab();
    await this.page.getByTestId("cms-article-add").click();
    const modal = await waitForDialogReady(this.articleModal);
    await modal.getByTestId("cms-article-title-input").fill(input.title);
    await modal.getByTestId("cms-article-slug-input").fill(input.slug);
    await this.uploadArticleCover(modal);
    await this.fillArticleContent(modal, input.content);
    await this.selectOptionByText(
      modal.getByTestId("cms-article-category-input"),
      new RegExp(input.categoryName, "i"),
    );
    await this.selectOptionByText(
      modal.getByTestId("cms-article-status-input"),
      /已发布|Published/i,
    );
    await this.confirmDialog(modal);
    await this.searchArticle(input.title);
  }

  async expectArticleEditorVisible() {
    await this.openArticlesTab();
    await this.page.getByTestId("cms-article-add").click();
    const modal = await waitForDialogReady(this.articleModal);
    await expect(modal.getByTestId("cms-article-cover-upload")).toBeVisible();
    await expect(modal.getByTestId("cms-article-content-editor")).toBeVisible();
    await expect(this.articleContentEditor(modal)).toBeVisible();
    await this.expectArticleEditorToolbar(modal);
    await this.expectArticleModalFitsViewport(modal);
    await modal.getByRole("button", { name: /取\s*消|Cancel/i }).last().click();
    await modal.waitFor({ state: "hidden", timeout: 10_000 }).catch(() => {});
  }

  async expectContentModulesVisible() {
    await this.openArticlesTab();
    await expect(this.page.getByTestId("cms-content-nav")).toContainText(
      /专题内容|Topic Content/i,
    );
    await expect(this.page.getByTestId("cms-content-nav")).toContainText(
      /新闻内容|News Content/i,
    );
  }

  async expectArticleListLayoutStable() {
    await this.openArticlesTab();
    await this.page.getByTestId("cms-content-model-model-list").click();
    await waitForBusyIndicatorsToClear(this.page);
    await expect
      .poll(async () =>
        this.page.evaluate(() => {
          const table = document.querySelector(
            '[data-testid="cms-article-table"]',
          ) as HTMLElement | null;
          const panel = document.querySelector(
            ".cms-content-layout > .cms-panel:last-child",
          ) as HTMLElement | null;
          const titleCell = table?.querySelector(
            "tbody tr td:first-child",
          ) as HTMLElement | null;
          const actionCell = table?.querySelector(
            "tbody tr td:last-child",
          ) as HTMLElement | null;
          const tableRect = table?.getBoundingClientRect();
          const panelRect = panel?.getBoundingClientRect();
          const titleRect = titleCell?.getBoundingClientRect();
          const actionRect = actionCell?.getBoundingClientRect();
          return {
            actionInsidePanel:
              !!actionRect &&
              !!panelRect &&
              actionRect.right <= panelRect.right + 1,
            tableFitsPanel:
              !!table &&
              !!panel &&
              !!tableRect &&
              !!panelRect &&
              table.scrollWidth <= Math.ceil(tableRect.width) + 1 &&
              tableRect.right <= panelRect.right + 1,
            titleWidth: Math.round(titleRect?.width ?? 0),
          };
        }),
      )
      .toEqual({
        actionInsidePanel: true,
        tableFitsPanel: true,
        titleWidth: expect.any(Number),
      });
    const titleWidth = await this.page
      .getByTestId("cms-article-table")
      .locator("tbody tr td:first-child")
      .first()
      .evaluate((node) => Math.round(node.getBoundingClientRect().width));
    expect(titleWidth).toBeLessThanOrEqual(420);
  }

  async expectArticleCategoryTreeCollapsible(input: {
    childId: number;
    childName: string;
    parentId: number;
    parentName: string;
  }) {
    await this.page.reload({ waitUntil: "domcontentloaded" });
    await this.expectReady();
    await this.openArticlesTab();
    const parent = this.page.getByTestId(
      `cms-content-category-${input.parentId}`,
    );
    const child = this.page.getByTestId(`cms-content-category-${input.childId}`);
    await expect(parent).toContainText(input.parentName);
    await expect(child).toHaveCount(0);

    const toggle = this.page.getByTestId(
      `cms-content-category-toggle-${input.parentId}`,
    );
    await expect(toggle).toHaveAttribute("aria-expanded", "false");
    await toggle.click();
    await expect(toggle).toHaveAttribute("aria-expanded", "true");
    await expect(child).toContainText(input.childName);
  }

  async expectModelContainsArticle(input: {
    model: "list" | "single";
    title: string;
  }) {
    await this.openArticlesTab();
    await this.page
      .getByTestId(`cms-content-model-model-${input.model}`)
      .click();
    await this.filterArticleTitle(input.title);
    await this.expectTableText("cms-article-table", input.title);
  }

  async expectCategoryGroupContainsArticle(input: {
    categoryId: number;
    title: string;
  }) {
    await this.openArticlesTab();
    await this.page
      .getByTestId(`cms-content-category-${input.categoryId}`)
      .click();
    await this.filterArticleTitle(input.title);
    await this.expectTableText("cms-article-table", input.title);
  }

  async expectArticleSourceEditor(input: {
    content: string;
    title: string;
  }) {
    await this.openArticlesTab();
    await this.searchArticle(input.title);
    const row = this.tableRowByText("cms-article-table", input.title);
    await row.waitFor({ state: "visible", timeout: 10_000 });
    await row.getByRole("button", { name: /编辑|Edit/i }).click();
    const modal = await waitForDialogReady(this.articleModal);
    const editor = modal.getByTestId("cms-article-content-editor");

    await editor.scrollIntoViewIfNeeded();
    const modeSwitch = editor.getByTestId("tiptap-editor-mode");
    await modeSwitch.scrollIntoViewIfNeeded();
    await modeSwitch.getByTestId("tiptap-mode-source").click();
    const sourceInput = editor.getByTestId("tiptap-source-input");
    await expect(sourceInput).toHaveValue(/<p>/);
    await expect(sourceInput).toHaveValue(new RegExp(input.content));
    await expect(sourceInput).not.toHaveValue(/&lt;p/);
    await expect(editor.getByTestId("tiptap-source-preview")).toContainText(
      input.content,
    );
    await modal.getByRole("button", { name: /取\s*消|Cancel/i }).last().click();
    await modal.waitFor({ state: "hidden", timeout: 10_000 }).catch(() => {});
  }

  async deleteArticle(title: string) {
    await this.openArticlesTab();
    await this.selectArticleModel("list");
    await this.filterArticleTitle(title);
    const row = this.tableRowByText("cms-article-table", title);
    await row.waitFor({ state: "visible", timeout: 10_000 });
    await row.getByRole("button", { name: /删除|Delete/i }).click();
    await this.confirmPopconfirm();
    await waitForRouteReady(this.page);
  }

  async searchArticle(title: string) {
    await this.openArticlesTab();
    await this.selectArticleModel("list");
    await this.filterArticleTitle(title);
    await this.expectTableText("cms-article-table", title);
  }

  async approveMessage(content: string) {
    await this.openMessagesTab();
    const row = this.tableRowByText("cms-message-table", content);
    await row.waitFor({ state: "visible", timeout: 10_000 });
    await row.getByRole("button", { name: /通过|Approve/i }).click();
    await this.confirmMessageModal();
    await waitForBusyIndicatorsToClear(this.page);
  }

  async expectClearAndLoadActionsRefreshAllSections() {
    await this.page.route("**/api/v1/cms/site/data", async (route) => {
      await route.fulfill({ contentType: "application/json", json: { code: 0 } });
    });
    await this.page.route("**/api/v1/cms/site/sample-data", async (route) => {
      await route.fulfill({ contentType: "application/json", json: { code: 0 } });
    });

    await this.openArticlesTab();
    await this.page.getByTestId("cms-article-title-filter").fill("stale filter");
    await this.page.getByTestId("cms-section-site").click();

    await this.expectSiteActionRefreshesAllSections("cms-site-clear-data", "DELETE");
    await expect(this.page.getByTestId("cms-article-title-filter")).toHaveValue("");

    await this.openSlidesTab();
    await this.page.getByTestId("cms-slide-keyword-filter").fill("stale slide");
    await this.page.getByTestId("cms-section-site").click();

    await this.expectSiteActionRefreshesAllSections(
      "cms-site-load-sample-data",
      "POST",
    );
    await this.openSlidesTab();
    await expect(this.page.getByTestId("cms-slide-keyword-filter")).toHaveValue("");
  }

  private tableRowByText(testId: string, text: string) {
    return this.page
      .getByTestId(testId)
      .locator("tbody tr", { hasText: text })
      .first();
  }

  private async expectTableText(testId: string, text: string) {
    await this.page
      .getByTestId(testId)
      .getByText(text, { exact: false })
      .first()
      .waitFor({
        state: "visible",
        timeout: 10_000,
      });
  }

  private async expectSiteActionRefreshesAllSections(
    testId: string,
    actionMethod: string,
  ) {
    const refreshPaths = [
      "/api/v1/cms/site",
      "/api/v1/cms/categories",
      "/api/v1/cms/articles",
      "/api/v1/cms/messages",
      "/api/v1/cms/slides",
      "/api/v1/cms/links",
    ];
    const refreshResponses = refreshPaths.map((path) =>
      this.page.waitForResponse(
        (response) =>
          response.url().includes(path) &&
          response.request().method() === "GET" &&
          response.status() === 200,
      ),
    );
    const actionResponse = this.page.waitForResponse(
      (response) =>
        response.url().includes(
          testId === "cms-site-clear-data"
            ? "/api/v1/cms/site/data"
            : "/api/v1/cms/site/sample-data",
        ) &&
        response.request().method() === actionMethod &&
        response.status() === 200,
    );

    await this.page.getByTestId(testId).click();
    await this.confirmPopconfirm();
    await actionResponse;
    await Promise.all(refreshResponses);
    await waitForBusyIndicatorsToClear(this.page);
    await this.openArticlesTab();
    await expect(this.page.getByTestId("cms-content-model-model-list")).toHaveClass(
      /is-active/,
    );
  }

  private async selectArticleModel(model: "list" | "single") {
    await this.page.getByTestId(`cms-content-model-model-${model}`).click();
    await waitForBusyIndicatorsToClear(this.page);
  }

  private async filterArticleTitle(title: string) {
    await this.page.getByTestId("cms-article-title-filter").fill(title);
    await this.page.getByTestId("cms-article-query").click();
    await waitForBusyIndicatorsToClear(this.page);
  }

  private articleContentEditor(dialog: Locator) {
    return dialog
      .getByTestId("cms-article-content-editor")
      .locator('[contenteditable="true"]')
      .first();
  }

  private async fillArticleContent(dialog: Locator, content: string) {
    const editor = this.articleContentEditor(dialog);
    await editor.waitFor({ state: "visible", timeout: 5000 });
    await editor.fill(content);
    await expect(editor).toContainText(content);
  }

  private async expectArticleEditorToolbar(dialog: Locator) {
    const editor = dialog.getByTestId("cms-article-content-editor");
    await expect(editor.getByTestId("tiptap-editor-mode")).toBeVisible();
    await expect(editor.getByTestId("tiptap-mode-source")).toBeVisible();
    await expect(editor.locator(".tiptap-toolbar")).toBeVisible();
    await expect(editor.locator(".tiptap-toolbar button")).toHaveCount(16);
  }

  private async expectArticleModalFitsViewport(dialog: Locator) {
    await expect
      .poll(async () =>
        dialog.evaluate((node) => {
          const body = node.querySelector(".ant-modal-body") as HTMLElement | null;
          const modalRect = node.getBoundingClientRect();
          const bodyRect = body?.getBoundingClientRect();
          return {
            bodyScrollable:
              !!body && body.scrollHeight >= Math.floor(body.clientHeight),
            bodyWithinViewport:
              !!bodyRect && bodyRect.bottom <= window.innerHeight - 64,
            modalWithinViewport: modalRect.bottom <= window.innerHeight + 2,
          };
        }),
      )
      .toEqual({
        bodyScrollable: true,
        bodyWithinViewport: true,
        modalWithinViewport: true,
      });
  }

  private async uploadArticleCover(dialog: Locator) {
    const upload = dialog.getByTestId("cms-article-cover-upload");
    await expect(upload).toBeVisible();
    const uploadResponse = this.page.waitForResponse(
      (response) =>
        response.url().includes("/api/v1/file/upload") &&
        response.request().method() === "POST" &&
        response.status() === 200,
    );
    await upload.locator('input[type="file"]').setInputFiles({
      name: "cms-cover.png",
      mimeType: "image/png",
      buffer: coverPngBuffer,
    });
    await uploadResponse;
    await waitForUploadReady(upload);
  }

  private async uploadSiteImage(testId: string, fileName: string) {
    const upload = this.page.getByTestId(testId);
    await expect(upload).toBeVisible();
    const uploadResponse = this.page.waitForResponse(
      (response) =>
        response.url().includes("/api/v1/file/upload") &&
        response.request().method() === "POST" &&
        response.status() === 200,
    );
    await upload.locator('input[type="file"]').setInputFiles({
      name: fileName,
      mimeType: "image/png",
      buffer: coverPngBuffer,
    });
    const response = await uploadResponse;
    const payload = (await response.json()) as {
      data?: { url?: string };
    };
    expect(payload.data?.url).toContain("/api/v1/uploads/");
    await waitForUploadReady(upload);
    await expect
      .poll(async () =>
        upload
          .locator(".ant-upload-list-item")
          .last()
          .evaluate((node) => {
            const image = node.querySelector("img");
            const anchor = node.querySelector("a");
            return `${image?.getAttribute("src") ?? ""} ${
              anchor?.getAttribute("href") ?? ""
            } ${node.textContent ?? ""}`;
          }),
      )
      .toContain("/api/v1/uploads/");
  }

  private async confirmDialog(dialog: Locator) {
    await dialog
      .getByRole("button", { name: /确\s*定|OK/i })
      .last()
      .click();
    await dialog.waitFor({ state: "hidden", timeout: 10_000 }).catch(() => {});
    await waitForRouteReady(this.page);
  }

  private async confirmPopconfirm() {
    const overlay = await waitForConfirmOverlay(this.page);
    await overlay
      .getByRole("button", { name: /确\s*定|OK|Yes/i })
      .last()
      .click();
    await overlay.waitFor({ state: "hidden", timeout: 10_000 }).catch(() => {});
    await waitForBusyIndicatorsToClear(this.page);
  }

  private async confirmModal() {
    const dialog = this.page.locator(".ant-modal-confirm:visible").last();
    await dialog.waitFor({ state: "visible", timeout: 5000 });
    await dialog
      .getByRole("button", { name: /确\s*定|OK/i })
      .last()
      .click();
    await dialog.waitFor({ state: "hidden", timeout: 10_000 }).catch(() => {});
  }

  private async confirmMessageModal() {
    const dialog = await waitForDialogReady(this.messageModal);
    await this.confirmDialog(dialog);
  }

  private async selectFirstOption(select: Locator) {
    await select.click();
    const option = this.page
      .locator(".ant-select-dropdown:visible .ant-select-item-option")
      .first();
    await option.waitFor({ state: "visible", timeout: 5000 });
    await option.click();
  }

  private async selectOptionByText(select: Locator, text: RegExp) {
    await select.click();
    const option = this.page
      .locator(".ant-select-dropdown:visible .ant-select-item-option", {
        hasText: text,
      })
      .first();
    await option.waitFor({ state: "visible", timeout: 5000 });
    await option.click();
  }
}
