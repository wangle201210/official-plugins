import { test, expect } from "@host-tests/fixtures/auth";
import { ensureSourcePluginEnabled } from "@host-tests/fixtures/plugin";
import { pgEscapeLiteral, queryPgScalar } from "@host-tests/support/postgres";

import { SicauNiuRecordPage } from "../pages/SicauNiuRecordPage";

const pluginID = "sicau-niu";
// Activation rows persist an opaque photo identifier, never a browser-resolvable
// URL: the bytes live in private plugin object storage and are only readable
// through the protected audit endpoint. The seeds below therefore store an
// identifier, and the tests stub that endpoint so the assertion proves the page
// exchanges the identifier for image bytes instead of rendering it as a URL.
const auditPhotoRoute =
  "**/plugins/sicau-niu/admin/audit/photos/*";
const auditPhotoBase64 = "R0lGODlhAQABAAAAACwAAAAAAQABAAA=";
const auditPhotoDataURL = `data:image/webp;base64,${auditPhotoBase64}`;

function photoIdentifier(suffix: string) {
  return `e2e-photo-${suffix}`;
}

function seedActivationRecordWithPhoto(suffix: string) {
  const openid = pgEscapeLiteral(`e2e-activation-photo-${suffix}`);
  const nickname = pgEscapeLiteral(`拍照玩家-${suffix}`);
  const niuCode = pgEscapeLiteral(`E2E-PHOTO-${suffix}`);
  const niuName = pgEscapeLiteral(`照片牛-${suffix}`);
  const photoPath = pgEscapeLiteral(photoIdentifier(suffix));
  return queryPgScalar(`
WITH inserted_user AS (
  INSERT INTO plugin_sicau_niu_user ("openid", "nickname")
  VALUES ('${openid}', '${nickname}')
  RETURNING "id"
),
inserted_niu AS (
  INSERT INTO plugin_sicau_niu_niu ("code", "niu_type", "name", "lat", "lng", "online_at", "status")
  VALUES ('${niuCode}', 'common', '${niuName}', 30.7035, 103.8290, CURRENT_TIMESTAMP - INTERVAL '1 day', 'active')
  RETURNING "id"
),
inserted_activation AS (
  INSERT INTO plugin_sicau_niu_activation (
    "user_id", "niu_id", "activity_date", "activated_at", "is_first", "order_no", "photo_path"
  )
  SELECT inserted_user."id", inserted_niu."id", CURRENT_DATE::text, CURRENT_TIMESTAMP, 1, 1, '${photoPath}'
  FROM inserted_user, inserted_niu
  RETURNING "id"
)
SELECT "id" FROM inserted_activation;
`);
}

function seedActivationAttemptWithPhoto(suffix: string) {
  const openid = pgEscapeLiteral(`e2e-activation-attempt-${suffix}`);
  const nickname = pgEscapeLiteral(`尝试玩家-${suffix}`);
  const niuCode = pgEscapeLiteral(`E2E-ATTEMPT-${suffix}`);
  const niuName = pgEscapeLiteral(`尝试牛-${suffix}`);
  const photoPath = pgEscapeLiteral(photoIdentifier(suffix));
  return queryPgScalar(`
WITH inserted_user AS (
  INSERT INTO plugin_sicau_niu_user ("openid", "nickname")
  VALUES ('${openid}', '${nickname}')
  RETURNING "id"
),
inserted_niu AS (
  INSERT INTO plugin_sicau_niu_niu ("code", "niu_type", "name", "lat", "lng", "online_at", "status")
  VALUES ('${niuCode}', 'common', '${niuName}', 30.7035, 103.8290, CURRENT_TIMESTAMP - INTERVAL '1 day', 'inactive')
  RETURNING "id"
),
inserted_attempt AS (
  INSERT INTO plugin_sicau_niu_activation_attempt (
    "user_id", "niu_id", "nearest_niu_id", "result", "lat", "lng", "distance_m", "threshold_m", "photo_path", "attempted_at"
  )
  SELECT inserted_user."id", 0, inserted_niu."id", 'out_of_range', 30.7135, 103.8290, 1111.9, 50, '${photoPath}', CURRENT_TIMESTAMP
  FROM inserted_user, inserted_niu
  RETURNING "id"
)
SELECT "id" FROM inserted_attempt;
`);
}

// TC-6 covers the sicau-niu operator activity-record query pages owned by the
// niu-activity-records change: the read-only feeding, steal and grass-ledger lists
// nested under the "寻牛记录" directory. It navigates through that group and asserts
// each page mounts and its read-only grid renders (a page-unique column header
// becomes visible), proving the page, route and list API are wired. Row presence
// depends on activity data, so the assertions anchor on the grid structure rather
// than a specific row.
test.describe("TC-6 sicau-niu 活动记录查询", () => {
  let recordPage: SicauNiuRecordPage;

  test.beforeEach(async ({ adminPage }) => {
    await ensureSourcePluginEnabled(adminPage, pluginID);
    recordPage = new SicauNiuRecordPage(adminPage);
  });

  test("TC-6a: 喂草记录页加载且表格渲染", async () => {
    await recordPage.openRecord("喂草记录");
    await recordPage.expectGridRendered("实际效果");
  });

  test("TC-6b: 偷草记录页加载且表格渲染", async () => {
    await recordPage.openRecord("偷草记录");
    await recordPage.expectGridRendered("被偷玩家");
  });

  test("TC-6c: 草账户流水页加载且表格渲染", async () => {
    await recordPage.openRecord("草账户流水");
    await recordPage.expectGridRendered("增减量");
  });

  test("TC-6d: 激活记录可查看上传照片", async ({ adminPage }) => {
    const suffix = `${Date.now()}`;
    const activationId = seedActivationRecordWithPhoto(suffix);
    const requestedPhotoIds: string[] = [];
    await adminPage.route(auditPhotoRoute, async (route) => {
      requestedPhotoIds.push(
        decodeURIComponent(new URL(route.request().url()).pathname.split("/").pop() ?? ""),
      );
      await route.fulfill({
        contentType: "application/json",
        body: JSON.stringify({
          code: 0,
          message: "OK",
          data: {
            contentType: "image/webp",
            sizeBytes: 64,
            imageBase64: auditPhotoBase64,
          },
        }),
      });
    });

    await recordPage.openRecord("激活记录");
    await recordPage.expectGridRendered("照片");
    await recordPage.expectActivationPhotoPreview(
      activationId,
      auditPhotoDataURL,
    );
    // 预览必须由不透明标识换取图片字节，而不是把标识当作图片地址渲染。
    expect(requestedPhotoIds).toContain(photoIdentifier(suffix));
  });

  test("TC-6e: 打卡尝试记录可查看失败原因和照片", async ({ adminPage }) => {
    const suffix = `${Date.now()}`;
    const attemptId = seedActivationAttemptWithPhoto(suffix);
    await adminPage.route(auditPhotoRoute, async (route) => {
      await route.fulfill({
        contentType: "application/json",
        body: JSON.stringify({
          code: 0,
          message: "OK",
          data: {
            contentType: "image/webp",
            sizeBytes: 64,
            imageBase64: auditPhotoBase64,
          },
        }),
      });
    });

    await recordPage.openRecord("激活记录");
    await recordPage.page.getByRole("tab", { name: "打卡尝试" }).click();
    await recordPage.expectGridRendered("判距(米)");
    await recordPage.expectActivationRecordPageHeightStable();
    await expect(recordPage.page.getByText("超出判距").first()).toBeVisible();
    await recordPage.expectActivationAttemptPhotoPreview(
      attemptId,
      auditPhotoDataURL,
    );
  });

  test("TC-6f: 激活照片读取失败展示可恢复提示", async ({ adminPage }) => {
    const suffix = `${Date.now()}`;
    const activationId = seedActivationRecordWithPhoto(suffix);
    await adminPage.route(auditPhotoRoute, async (route) => {
      await route.fulfill({
        status: 404,
        contentType: "application/json",
        body: JSON.stringify({
          code: 50000,
          message: "Activation photo not found",
        }),
      });
    });

    await recordPage.openRecord("激活记录");
    await recordPage.expectGridRendered("照片");
    await recordPage.expectActivationPhotoPreviewFailure(
      activationId,
      "照片读取失败,请稍后重试",
    );
  });
});
