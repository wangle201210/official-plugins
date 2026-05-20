import { expect, type Locator, type Page } from "@host-tests/support/playwright";

import { PluginPage } from "@host-tests/pages/PluginPage";

export class DemoDynamicPage extends PluginPage {
  constructor(page: Page) {
    super(page);
  }

  pluginDemoDynamicTitle(): Locator {
    return this.page.getByTestId("linapro-demo-dynamic-title").first();
  }

  pluginDemoDynamicDescription(): Locator {
    return this.page.getByTestId("linapro-demo-dynamic-description").first();
  }

  pluginDemoDynamicOpenStandaloneButton(): Locator {
    return this.page.getByTestId("linapro-demo-dynamic-open-standalone").first();
  }

  pluginDemoDynamicRecordGrid(): Locator {
    return this.page.getByTestId("linapro-demo-dynamic-record-grid").first();
  }

  pluginDemoDynamicRecordAddButton(): Locator {
    return this.page.getByTestId("linapro-demo-dynamic-record-add").first();
  }

  pluginDemoDynamicRecordPagination(): Locator {
    return this.page
      .getByTestId("linapro-demo-dynamic-record-pagination")
      .first();
  }

  pluginDemoDynamicPaginationSummary(): Locator {
    return this.page
      .getByTestId("linapro-demo-dynamic-pagination-summary")
      .first();
  }

  pluginDemoDynamicPaginationPage(pageNumber: number): Locator {
    return this.page
      .getByTestId(`linapro-demo-dynamic-pagination-page-${pageNumber}`)
      .first();
  }

  pluginDemoDynamicPaginationPrevButton(): Locator {
    return this.page.getByTestId("linapro-demo-dynamic-pagination-prev").first();
  }

  pluginDemoDynamicPaginationNextButton(): Locator {
    return this.page.getByTestId("linapro-demo-dynamic-pagination-next").first();
  }

  pluginDemoDynamicRecordModal(): Locator {
    return this.page.getByTestId("linapro-demo-dynamic-record-modal").last();
  }

  pluginDemoDynamicRecordTitleInput(): Locator {
    return this.page
      .getByTestId("linapro-demo-dynamic-record-title-input")
      .last();
  }

  pluginDemoDynamicRecordContentInput(): Locator {
    return this.page
      .getByTestId("linapro-demo-dynamic-record-content-input")
      .last();
  }

  pluginDemoDynamicRecordFileInput(): Locator {
    return this.page
      .getByTestId("linapro-demo-dynamic-record-file-input")
      .last();
  }

  pluginDemoDynamicRecordRemoveAttachment(): Locator {
    return this.page
      .getByTestId("linapro-demo-dynamic-record-remove-attachment")
      .last();
  }

  pluginDemoDynamicRecordSubmitButton(): Locator {
    return this.page.getByTestId("linapro-demo-dynamic-record-submit").last();
  }

  pluginDemoDynamicRecordRow(title: string): Locator {
    return this.pluginDemoDynamicRecordGrid()
      .locator("tbody tr", { hasText: title })
      .first();
  }

  pluginDemoDynamicEditButton(title: string): Locator {
    return this.pluginDemoDynamicRecordRow(title)
      .getByRole("button", { name: "编辑" })
      .first();
  }

  pluginDemoDynamicDeleteButton(title: string): Locator {
    return this.pluginDemoDynamicRecordRow(title)
      .getByRole("button", { name: "删除" })
      .first();
  }

  async createPluginDemoDynamicRecord(input: {
    attachmentPath?: string;
    content: string;
    title: string;
  }) {
    await this.pluginDemoDynamicRecordAddButton().click();
    await expect(this.pluginDemoDynamicRecordModal()).toBeVisible();
    await this.pluginDemoDynamicRecordTitleInput().fill(input.title);
    await this.pluginDemoDynamicRecordContentInput().fill(input.content);
    if (input.attachmentPath) {
      await this.pluginDemoDynamicRecordFileInput().setInputFiles(
        input.attachmentPath,
      );
    }
    await this.pluginDemoDynamicRecordSubmitButton().click();
    await expect(this.pluginDemoDynamicRecordModal()).toHaveAttribute(
      "data-open",
      "false",
    );
    await expect(this.pluginDemoDynamicRecordRow(input.title)).toBeVisible();
  }

  async updatePluginDemoDynamicRecord(
    currentTitle: string,
    input: {
      attachmentPath?: string;
      content: string;
      removeAttachment?: boolean;
      title: string;
    },
  ) {
    await this.pluginDemoDynamicEditButton(currentTitle).click();
    await expect(this.pluginDemoDynamicRecordModal()).toBeVisible();
    await this.pluginDemoDynamicRecordTitleInput().fill(input.title);
    await this.pluginDemoDynamicRecordContentInput().fill(input.content);
    if (input.removeAttachment) {
      const checkbox = this.pluginDemoDynamicRecordRemoveAttachment().locator(
        'input[type="checkbox"]',
      );
      if ((await checkbox.isChecked()) !== true) {
        await checkbox.click();
      }
    }
    if (input.attachmentPath) {
      await this.pluginDemoDynamicRecordFileInput().setInputFiles(
        input.attachmentPath,
      );
    }
    await this.pluginDemoDynamicRecordSubmitButton().click();
    await expect(this.pluginDemoDynamicRecordModal()).toHaveAttribute(
      "data-open",
      "false",
    );
    await expect(this.pluginDemoDynamicRecordRow(input.title)).toBeVisible();
  }

  async deletePluginDemoDynamicRecord(title: string) {
    this.page.once("dialog", async (dialog) => {
      await dialog.accept();
    });
    await this.pluginDemoDynamicDeleteButton(title).click();
    await expect(this.pluginDemoDynamicRecordRow(title)).toHaveCount(0);
  }
}
