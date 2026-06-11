import { test, expect } from "@host-tests/fixtures/auth";
import { ensureSourcePluginEnabled } from "@host-tests/fixtures/plugin";
import { pgEscapeLiteral, queryPgScalar } from "@host-tests/support/postgres";

import { SicauNiuRecordPage } from "../pages/SicauNiuRecordPage";

const pluginID = "sicau-niu";
const tinyPhotoDataURL =
  "data:image/gif;base64,R0lGODlhAQABAAAAACwAAAAAAQABAAA=";

function seedActivationRecordWithPhoto(suffix: string) {
  const openid = pgEscapeLiteral(`e2e-activation-photo-${suffix}`);
  const nickname = pgEscapeLiteral(`拍照玩家-${suffix}`);
  const niuCode = pgEscapeLiteral(`E2E-PHOTO-${suffix}`);
  const niuName = pgEscapeLiteral(`照片牛-${suffix}`);
  const photoPath = pgEscapeLiteral(tinyPhotoDataURL);
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
  const photoPath = pgEscapeLiteral(tinyPhotoDataURL);
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
// nested under the "活动记录" sub-directory. It navigates two levels deep and asserts
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

  test("TC-6d: 激活记录可查看上传照片", async () => {
    const activationId = seedActivationRecordWithPhoto(`${Date.now()}`);

    await recordPage.openRecord("激活记录");
    await recordPage.expectGridRendered("照片");
    await recordPage.expectActivationPhotoPreview(
      activationId,
      tinyPhotoDataURL,
    );
  });

  test("TC-6e: 打卡尝试记录可查看失败原因和照片", async () => {
    const attemptId = seedActivationAttemptWithPhoto(`${Date.now()}`);

    await recordPage.openRecord("激活记录");
    await recordPage.page.getByRole("tab", { name: "打卡尝试" }).click();
    await recordPage.expectGridRendered("判距(米)");
    await recordPage.expectActivationRecordPageHeightStable();
    await expect(recordPage.page.getByText("超出判距").first()).toBeVisible();
    await recordPage.expectActivationAttemptPhotoPreview(
      attemptId,
      tinyPhotoDataURL,
    );
  });
});
