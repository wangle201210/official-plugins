import { test, expect } from '../../support/linapro-tenant-core';
import { scenarioTC0181 } from '../../support/linapro-tenant-core-scenarios';

test.describe('TC-4 租户不暴露归档生命周期', () => {
  test.use({ multiTenantMode: 'linapro-tenant-core-enabled' });

  test('TC-4a: archived status transitions are rejected', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('linapro-tenant-core-enabled');
    await scenarioTC0181();
  });
});
