import { test, expect } from '../../support/linapro-tenant-core';
import { scenarioTC0215 } from '../../support/linapro-tenant-core-scenarios';

test.describe('TC-3 impersonation 审计日志', () => {
  test.use({ multiTenantMode: 'linapro-tenant-core-enabled' });

  test('TC-3a: impersonation writes dual-track audit fields', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('linapro-tenant-core-enabled');
    await scenarioTC0215();
  });
});
