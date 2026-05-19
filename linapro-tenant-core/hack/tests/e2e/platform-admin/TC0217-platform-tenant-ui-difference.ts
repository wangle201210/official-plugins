import { test, expect } from '../../support/linapro-tenant-core';
import { scenarioTC0217 } from '../../support/linapro-tenant-core-scenarios';

test.describe('TC-217 平台与租户视图差异', () => {
  test.use({ multiTenantMode: 'linapro-tenant-core-enabled' });

  test('TC-217a: platform and tenant API menu views differ', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('linapro-tenant-core-enabled');
    await scenarioTC0217();
  });
});
