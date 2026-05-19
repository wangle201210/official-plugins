import { test, expect } from '../../support/linapro-tenant-core';
import { scenarioTC0202 } from '../../support/linapro-tenant-core-scenarios';

test.describe('TC-202 default 解析器 ambiguous prompt', () => {
  test.use({ multiTenantMode: 'linapro-tenant-core-enabled' });

  test('TC-202a: ambiguous login returns preToken and tenant candidates', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('linapro-tenant-core-enabled');
    await scenarioTC0202();
  });
});
