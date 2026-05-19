import { test, expect } from '../../support/linapro-tenant-core';
import { scenarioTC0239 } from '../../support/linapro-tenant-core-scenarios';

test.describe('TC-239 租户角色平台权限阻断', () => {
  test.use({ multiTenantMode: 'linapro-tenant-core-enabled' });

  test('TC-239a: dirty platform grants are hidden and rejected in tenant context', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('linapro-tenant-core-enabled');
    await scenarioTC0239();
  });
});
