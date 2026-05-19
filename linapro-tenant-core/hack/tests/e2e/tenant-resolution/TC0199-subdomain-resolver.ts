import { test, expect } from '../../support/linapro-tenant-core';
import { scenarioTC0199 } from '../../support/linapro-tenant-core-scenarios';

test.describe('TC-199 subdomain 解析器', () => {
  test.use({ multiTenantMode: 'linapro-tenant-core-enabled' });

  test('TC-199a: subdomain root is fixed empty and reserved labels are enforced', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('linapro-tenant-core-enabled');
    await scenarioTC0199();
  });
});
