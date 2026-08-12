import { expect, test } from "@host-tests/fixtures/auth";
import { ensureSourcePluginEnabled } from "@host-tests/fixtures/plugin";
import { execPgSQL, pgEscapeLiteral } from "@host-tests/support/postgres";

import { SicauNiuTransportPage } from "../pages/SicauNiuTransportPage";

const pluginID = "sicau-niu";
const contributionMeters = 111;
const seededSuffixes: string[] = [];

function seedTransportFacts(suffix: string) {
  seededSuffixes.push(suffix);
  const teamName = `云搬牛验证团-${suffix}`;
  const openid = pgEscapeLiteral(`e2e-transport-${suffix}`);
  const nickname = pgEscapeLiteral(`搬牛玩家-${suffix}`);
  const escapedTeamName = pgEscapeLiteral(teamName);
  execPgSQL(`
WITH inserted_user AS (
  INSERT INTO plugin_sicau_niu_user ("openid", "nickname")
  VALUES ('${openid}', '${nickname}')
  RETURNING "id"
),
inserted_team AS (
  INSERT INTO plugin_sicau_niu_transport_team (
    "name", "leader_user_id", "create_request_id", "status", "visible",
    "member_count", "total_contribution_meters", "last_active_at"
  )
  SELECT '${escapedTeamName}', inserted_user."id", 'e2e-create-${suffix}',
    'effective', 1, 1, ${contributionMeters}, CURRENT_TIMESTAMP
  FROM inserted_user
  RETURNING "id", "leader_user_id"
),
inserted_member AS (
  INSERT INTO plugin_sicau_niu_transport_member (
    "team_id", "user_id", "join_request_id", "role", "total_contribution_meters",
    "last_report_lat", "last_report_lng", "last_report_at"
  )
  SELECT inserted_team."id", inserted_team."leader_user_id", 'e2e-join-${suffix}',
    'creator', ${contributionMeters}, 30.7068, 103.8318, CURRENT_TIMESTAMP
  FROM inserted_team
  RETURNING "id", "team_id", "user_id"
),
first_report AS (
  INSERT INTO plugin_sicau_niu_transport_report (
    "team_id", "member_id", "user_id", "request_id", "activity_date",
    "end_lat", "end_lng", "sampled_at", "accepted_at", "contribution_meters",
    "user_total_meters", "team_total_meters", "daily_report_count"
  )
  SELECT inserted_member."team_id", inserted_member."id", inserted_member."user_id",
    'e2e-report-1-${suffix}', CURRENT_DATE::text, 30.7058, 103.8318,
    CURRENT_TIMESTAMP - INTERVAL '5 minutes', CURRENT_TIMESTAMP - INTERVAL '5 minutes',
    0, 0, 0, 1
  FROM inserted_member
  RETURNING "id"
),
second_report AS (
  INSERT INTO plugin_sicau_niu_transport_report (
    "team_id", "member_id", "user_id", "request_id", "activity_date",
    "start_lat", "start_lng", "end_lat", "end_lng", "sampled_at", "accepted_at",
    "contribution_meters", "user_total_meters", "team_total_meters", "daily_report_count"
  )
  SELECT inserted_member."team_id", inserted_member."id", inserted_member."user_id",
    'e2e-report-2-${suffix}', CURRENT_DATE::text, 30.7058, 103.8318,
    30.7068, 103.8318, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP,
    ${contributionMeters}, ${contributionMeters}, ${contributionMeters}, 2
  FROM inserted_member, first_report
  RETURNING "id"
)
SELECT "id" FROM inserted_team;
`);
  return teamName;
}

function cleanTransportFacts() {
  for (const suffix of seededSuffixes.splice(0)) {
    const escapedSuffix = pgEscapeLiteral(suffix);
    execPgSQL(`
DELETE FROM plugin_sicau_niu_transport_report WHERE "request_id" IN ('e2e-report-1-${escapedSuffix}', 'e2e-report-2-${escapedSuffix}');
DELETE FROM plugin_sicau_niu_transport_member WHERE "join_request_id" = 'e2e-join-${escapedSuffix}';
DELETE FROM plugin_sicau_niu_transport_team WHERE "create_request_id" = 'e2e-create-${escapedSuffix}';
DELETE FROM plugin_sicau_niu_user WHERE "openid" = 'e2e-transport-${escapedSuffix}';
`);
  }
}

// TC-8 covers the cloud-moving operator surface: persisted aggregate facts,
// effective-team management, raw report audit and the operator-only rename path.
test.describe("TC-8 sicau-niu 云搬牛管理", () => {
  let transportPage: SicauNiuTransportPage;

  test.beforeEach(async ({ adminPage }) => {
    await ensureSourcePluginEnabled(adminPage, pluginID);
    transportPage = new SicauNiuTransportPage(adminPage);
    await transportPage.openFromMenu();
  });

  test.afterEach(() => {
    cleanTransportFacts();
  });

  test("TC-8a: 有效团与数据库贡献统计可见", async ({}, testInfo) => {
    const teamName = seedTransportFacts(`stats-${Date.now()}`);
    await transportPage.page.reload();
    await expect(transportPage.stats()).toContainText(
      String(contributionMeters),
    );
    await transportPage.expectSeededTeam(teamName, contributionMeters);
    await transportPage.page.screenshot({
      fullPage: true,
      path: testInfo.outputPath("cloud-moving-teams.png"),
    });
  });

  test("TC-8b: 管理员只有受保护入口可以修改团名", async () => {
    const suffix = `rename-${Date.now()}`;
    const teamName = seedTransportFacts(suffix);
    const renamedTeamName = `云搬牛已改名-${suffix}`;
    await transportPage.page.reload();
    await transportPage.renameTeam(teamName, renamedTeamName);
  });

  test("TC-8c: 有效团改名不能与其他有效团重名", async () => {
    const suffix = `duplicate-name-${Date.now()}`;
    const existingName = seedTransportFacts(`${suffix}-existing`);
    const targetName = seedTransportFacts(`${suffix}-target`);
    await transportPage.page.reload();
    await transportPage.expectDuplicateRenameRejected(targetName, existingName);
    await transportPage.page.reload();
    await expect(transportPage.teamRow(targetName)).toBeVisible();
    await expect(transportPage.teamRow(existingName)).toBeVisible();
  });

  test("TC-8d: 首次零贡献与后续位置事实均可审计", async ({}, testInfo) => {
    const teamName = seedTransportFacts(`audit-${Date.now()}`);
    await transportPage.page.reload();
    await transportPage.openReportAudit();
    await transportPage.expectReportFacts(teamName, contributionMeters);
    await transportPage.page.screenshot({
      fullPage: true,
      path: testInfo.outputPath("cloud-moving-report-audit.png"),
    });
  });
});
