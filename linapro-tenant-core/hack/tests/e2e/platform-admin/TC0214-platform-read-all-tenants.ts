import { test, expect } from '../../support/linapro-tenant-core';
import { scenarioTC0214 } from '../../support/linapro-tenant-core-scenarios';

test.describe('TC-214 平台管理员跨租户读', () => {
  test.use({ multiTenantMode: 'linapro-tenant-core-enabled' });

  test('TC-214a: platform tenant list can read tenants across scopes', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('linapro-tenant-core-enabled');
    await scenarioTC0214();
  });
});
