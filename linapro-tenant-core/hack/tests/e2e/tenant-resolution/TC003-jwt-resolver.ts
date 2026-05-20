import { test, expect } from '../../support/linapro-tenant-core';
import { scenarioTC0200 } from '../../support/linapro-tenant-core-scenarios';

test.describe('TC-3 jwt 解析器', () => {
  test.use({ multiTenantMode: 'linapro-tenant-core-enabled' });

  test('TC-3a: JWT tenant claim authorizes tenant member access', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('linapro-tenant-core-enabled');
    await scenarioTC0200();
  });
});
