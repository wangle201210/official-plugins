import { expect, test } from "@host-tests/fixtures/auth";
import { ensureSourcePluginEnabled } from "@host-tests/fixtures/plugin";
import {
  createAdminApiContext,
  expectSuccess,
} from "@host-tests/support/api/job";
import { workspacePath } from "@host-tests/fixtures/config";
import { waitForRouteReady } from "@host-tests/support/ui";

type CreatedId = {
  id: number;
};

type StrategyDetail = {
  id: number;
};

type WaterPreviewResult = {
  success: boolean;
  status: string;
  image: string;
  source: string;
  sourceLabel: string;
  strategyId: number;
  strategyName: string;
};

type WaterSubmitResult = {
  success: boolean;
  taskId: string;
  status: string;
};

type WaterTaskResult = {
  success: boolean;
  status: string;
  image: string;
  strategyId: number;
};

const testImage =
  "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAHgAAABQCAIAAABd+SbeAAAA0UlEQVR4nOzQQQ3AIADAwGWZjmlCIlJRQXlwp6Dp94/5sN97OuAWRkeMjhgdMTpidMToiNERoyNGR4yOGB0xOmJ0xOiI0RGjI0ZHjI4YHTE6YnTE6IjREaMjRkeMjhgdMTpidMToiNERoyNGR4yOGB0xOmJ0xOiI0RGjI0ZHjI4YHTE6YnTE6IjREaMjRkeMjhgdMTpidMToiNERoyNGR4yOGB0xOmJ0xOiI0RGjI0ZHjI4YHTE6YnTE6IjREaMjRkeMjhgdMTpidMToyAoAAP//BiYBsdlbvXQAAAAASUVORK5CYII=";

async function createWatermarkStrategy(api: Awaited<ReturnType<typeof createAdminApiContext>>) {
  const created = await expectSuccess<CreatedId>(
    await api.post("media/strategies", {
      data: {
        name: `水印E2E策略-${Date.now()}`,
        enable: 1,
        global: 0,
        strategy: `snapshot_watermark:
  enabled: true
  text: LinaPro Water
  fontSize: 18
  color: "#ffffff"
  align: bottomRight
  opacity: 0.8`,
      },
    }),
  );
  return created.id;
}

async function currentGlobalStrategyIds(api: Awaited<ReturnType<typeof createAdminApiContext>>) {
  const result = await expectSuccess<{ list: StrategyDetail[]; total: number }>(
    await api.get("media/strategies?pageNum=1&pageSize=20&global=1"),
  );
  return result.list.map((item) => item.id);
}

test.describe("TC-1 water source plugin", () => {
  test("TC-1a: water APIs render watermark from media_* strategy tables", async ({
    adminPage,
  }) => {
    await ensureSourcePluginEnabled(adminPage, "media");
    await ensureSourcePluginEnabled(adminPage, "water");

    const api = await createAdminApiContext();
    const deviceId = `water-e2e-device-${Date.now()}`;
    let deviceStrategyId = 0;
    let strategyId = 0;
    let previousGlobalIds: number[] = [];
    try {
      deviceStrategyId = await createWatermarkStrategy(api);
      await expectSuccess(
        await api.put(`media/device-bindings/${encodeURIComponent(deviceId)}`, {
          data: {
            deviceId,
            strategyId: deviceStrategyId,
          },
        }),
      );

      const devicePreview = await expectSuccess<WaterPreviewResult>(
        await api.post("water/preview", {
          data: {
            tenant: "tenant-water-e2e",
            deviceId,
            deviceCode: deviceId,
            channelCode: deviceId,
            image: testImage,
          },
        }),
      );
      expect(devicePreview.success).toBeTruthy();
      expect(devicePreview.status).toBe("success");
      expect(devicePreview.source).toBe("device");
      expect(devicePreview.sourceLabel).toBe("设备策略");
      expect(devicePreview.strategyId).toBe(deviceStrategyId);
      expect(devicePreview.image).toContain("data:image/png;base64,");

      previousGlobalIds = await currentGlobalStrategyIds(api);
      strategyId = await createWatermarkStrategy(api);
      await expectSuccess(await api.put(`media/strategies/${strategyId}/global`));

      const preview = await expectSuccess<WaterPreviewResult>(
        await api.post("water/preview", {
          data: {
            tenant: "tenant-water-e2e",
            deviceId: "34020000001320009999",
            image: testImage,
          },
        }),
      );
      expect(preview.success).toBeTruthy();
      expect(preview.status).toBe("success");
      expect(preview.source).toBe("global");
      expect(preview.sourceLabel).toBe("全局策略");
      expect(preview.strategyId).toBe(strategyId);
      expect(preview.image).toContain("data:image/png;base64,");

      const submitted = await expectSuccess<WaterSubmitResult>(
        await api.post("water/snaps/gb/34020000001320009999", {
          data: {
            tenant: "tenant-water-e2e",
            deviceCode: "34020000001320009999",
            image: testImage,
          },
        }),
      );
      expect(submitted.success).toBeTruthy();
      expect(submitted.status).toBe("queued");

      await expect
        .poll(
          async () => {
            const current = await expectSuccess<WaterTaskResult>(
              await api.get(`water/tasks/${submitted.taskId}`),
            );
            return `${current.success}:${current.status}:${current.strategyId}`;
          },
          { timeout: 10000 },
        )
        .toBe(`true:success:${strategyId}`);
      const task = await expectSuccess<WaterTaskResult>(
        await api.get(`water/tasks/${submitted.taskId}`),
      );
      expect(task.image).toContain("data:image/png;base64,");
    } finally {
      if (deviceStrategyId > 0) {
        await api.delete(`media/device-bindings/${encodeURIComponent(deviceId)}`).catch(() => {});
      }
      if (previousGlobalIds.length > 0) {
        await api.put(`media/strategies/${previousGlobalIds[0]}/global`).catch(() => {});
      }
      if (strategyId > 0) {
        await api.delete(`media/strategies/${strategyId}`).catch(() => {});
      }
      if (deviceStrategyId > 0) {
        await api.delete(`media/strategies/${deviceStrategyId}`).catch(() => {});
      }
      await api.dispose();
    }
  });

  test("TC-1b: water page opens without frontend exceptions", async ({
    adminPage,
  }) => {
    await ensureSourcePluginEnabled(adminPage, "water");
    const pageErrors: Error[] = [];
    adminPage.on("pageerror", (error) => pageErrors.push(error));

    await adminPage.goto(workspacePath("/water"), { waitUntil: "domcontentloaded" });
    await waitForRouteReady(adminPage, 15000);

    await expect(adminPage.getByRole("heading", { name: "水印服务" })).toBeVisible();
    await expect(adminPage.getByText("基于媒体策略表解析租户、设备和全局策略")).toBeVisible();
    expect(pageErrors.map((error) => error.message)).toEqual([]);
  });
});
