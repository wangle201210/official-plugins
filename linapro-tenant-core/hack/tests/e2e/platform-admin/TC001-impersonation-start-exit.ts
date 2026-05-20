import { test, expect } from '../../support/linapro-tenant-core';
import { scenarioTC0186 } from '../../support/linapro-tenant-core-scenarios';

test.describe('TC-1 平台管理员 impersonation', () => {
  test.use({ multiTenantMode: 'linapro-tenant-core-enabled' });

  test('TC-1a: platform impersonation starts and then revokes its online session', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('linapro-tenant-core-enabled');
    await scenarioTC0186();
  });
});
