-- 003: CMS plugin starter site content
-- 003：CMS 插件默认站点内容
--
-- This starter dataset is part of normal plugin installation. It gives a newly
-- installed CMS plugin a usable public site without requiring optional mock
-- data. The statements are idempotent and avoid deleting or replacing rows
-- that operators have already maintained after installation.

INSERT INTO plugin_cms_site ("site_key", "name", "logo", "weixin", "domain", "slogan", "keywords", "description", "icp", "contact", "phone", "email", "address", "status", "show_messages", "created_at", "updated_at")
SELECT 'default', '启明先进材料产业研究院', '/static/logo.svg', '/static/wechat.jpg', 'www.advanced-materials-demo.com', '启明先进材料产业研究院', '启明先进材料产业研究院', '启明先进材料产业研究院', '', '贾老师', '0731-88886666', 'admin@advanced-materials-demo.com', '示范区科创园先进材料公共服务中心', 1, 1, '2026-05-09 13:37:35', '2026-05-09 13:37:35'
WHERE NOT EXISTS (
    SELECT 1 FROM plugin_cms_site WHERE "site_key" = 'default'
);

UPDATE plugin_cms_site
SET
    "name" = '启明先进材料产业研究院',
    "logo" = '/static/logo.svg',
    "weixin" = '/static/wechat.jpg',
    "domain" = 'www.advanced-materials-demo.com',
    "slogan" = '启明先进材料产业研究院',
    "keywords" = '启明先进材料产业研究院',
    "description" = '启明先进材料产业研究院',
    "contact" = '贾老师',
    "phone" = '0731-88886666',
    "email" = 'admin@advanced-materials-demo.com',
    "address" = '示范区科创园先进材料公共服务中心',
    "status" = 1,
    "show_messages" = 1
WHERE "site_key" = 'default'
  AND "updated_by" = 0
  AND "name" = 'LinaPro CMS'
  AND "slogan" = 'AI-native full-stack delivery framework'
  AND "keywords" = 'LinaPro,CMS,AI-native'
  AND "description" = 'LinaPro CMS demo site';

INSERT INTO plugin_cms_category ("code", "parent_id", "name", "type", "path", "cover", "outlink", "title", "keywords", "description", "sort", "status", "created_at", "updated_at") VALUES
('1', 0, '关于我们', 2, '/about_1/', '', '', '', '', '', 255, 1, '2018-04-11 17:26:11', '2026-04-18 22:08:10'),
('2', 0, '党建与文化', 1, '/list_2/', '', '', '', '', '', 255, 1, '2018-04-11 17:26:46', '2026-04-18 22:08:10'),
('3', 0, '党建工作', 1, '/list_3/', '', '', '', '', '', 255, 1, '2018-04-11 17:27:05', '2026-04-18 22:08:10'),
('4', 0, '文化建设', 2, '/about_4/', '', '', '', '', '', 255, 1, '2018-04-11 17:27:30', '2026-04-18 22:08:10'),
('5', 0, '新闻中心', 1, '/list_5/', '', '', '', '', '', 255, 1, '2018-04-11 17:27:54', '2026-04-18 22:08:10'),
('12', 0, '建设背景', 2, '/about_12/', '', '', '', '', '', 255, 1, '2026-04-16 16:56:52', '2026-04-18 22:08:10'),
('13', 0, '研究院概况', 2, '/about_13/', '', '', '', '', '', 255, 1, '2026-04-18 15:35:49', '2026-04-18 22:08:10'),
('14', 0, '组织机构', 2, '/about_14/', '', '', '', '', '', 255, 1, '2026-04-18 15:37:31', '2026-04-18 22:08:10'),
('15', 0, '领导队伍', 1, '/list_15/', '', '', '', '', '', 255, 1, '2026-04-18 15:38:36', '2026-04-30 10:19:35'),
('16', 0, '联系我们', 2, '/about_16/', '', '', '', '', '', 255, 1, '2026-04-18 15:41:11', '2026-04-18 22:08:10'),
('17', 0, '院区风光', 1, '/list_17/', '', '', '', '', '', 255, 1, '2026-04-18 15:45:00', '2026-04-18 22:08:10'),
('18', 0, '品牌形象', 1, '/list_18/', '', '', '', '', '', 255, 1, '2026-04-18 15:45:00', '2026-04-18 22:08:10'),
('19', 0, '工作动态', 1, '/list_19/', '', '', '', '', '', 255, 1, '2026-04-18 15:46:28', '2026-04-18 22:08:10'),
('20', 0, '通知公告', 1, '/list_20/', '', '', '', '', '', 255, 1, '2026-04-18 15:46:28', '2026-04-18 22:08:10'),
('21', 0, '政策资讯', 1, '/list_21/', '', '', '', '', '', 255, 1, '2026-04-18 15:46:28', '2026-04-18 22:08:10'),
('22', 0, '媒体聚焦', 1, '/list_22/', '', '', '', '', '', 255, 1, '2026-04-18 15:46:28', '2026-04-18 22:08:10'),
('23', 0, '科技创新', 1, '/list_23/', '', '', '', '', '', 255, 1, '2026-04-18 15:46:55', '2026-04-18 22:08:10'),
('24', 0, '研究方向', 1, '/list_24/', '', '', '', '', '', 255, 1, '2026-04-18 15:47:27', '2026-04-18 22:08:10'),
('25', 0, '科研成果', 1, '/list_25/', '', '', '', '', '', 255, 1, '2026-04-18 15:47:27', '2026-04-18 22:08:10'),
('26', 0, '技术', 1, '/list_26/', '', '', '', '', '', 255, 1, '2026-04-18 15:48:14', '2026-04-18 22:08:10'),
('27', 0, '专利', 1, '/list_27/', '', '', '', '', '', 255, 1, '2026-04-18 15:48:14', '2026-04-18 22:08:10'),
('28', 0, '标准', 1, '/list_28/', '', '', '', '', '', 255, 1, '2026-04-18 15:48:14', '2026-04-18 22:08:10'),
('29', 0, '著作', 1, '/list_29/', '', '', '', '', '', 255, 1, '2026-04-18 15:48:14', '2026-04-18 22:08:10'),
('30', 0, '论文', 1, '/list_30/', '', '', '', '', '', 255, 1, '2026-04-18 15:48:14', '2026-04-18 22:08:10'),
('31', 0, '技术转移与成果转化', 1, '/list_31/', '', '', '', '', '', 255, 1, '2026-04-18 15:48:54', '2026-04-18 22:08:10'),
('32', 0, '技术转移', 1, '/list_32/', '', '', '', '', '', 255, 1, '2026-04-18 15:49:29', '2026-04-18 22:08:10'),
('33', 0, '成果转化', 1, '/list_33/', '', '', '', '', '', 255, 1, '2026-04-18 15:49:29', '2026-04-18 22:08:10'),
('34', 0, '共享平台', 1, '/list_34/', '', '', '', '', '', 255, 1, '2026-04-18 15:50:22', '2026-04-18 22:08:10'),
('35', 0, '实验室规章制度', 2, '/about_35/', '', '', '', '', '', 255, 1, '2026-04-18 15:51:02', '2026-04-18 22:08:10'),
('36', 0, '设备名录', 1, '/list_36/', '', '', '', '', '', 255, 1, '2026-04-18 15:51:02', '2026-04-18 22:08:10'),
('37', 0, '预约指南与流程', 2, '/about_37/', '', '', '', '', '', 255, 1, '2026-04-18 15:51:02', '2026-04-18 22:08:10'),
('38', 0, '收费标准', 2, '/about_38/', '', '', '', '', '', 255, 1, '2026-04-18 15:51:02', '2026-04-18 22:08:10'),
('39', 0, '常见问题', 2, '/about_39/', '', '', '', '', '', 255, 1, '2026-04-18 15:51:02', '2026-04-18 22:08:10'),
('40', 0, '产业服务', 2, '/about_40/', '', '', '', '', '', 255, 1, '2026-04-18 15:54:01', '2026-04-18 22:08:10'),
('42', 0, '经典案例', 1, '/list_42/', '', '', '', '', '', 255, 1, '2026-04-18 15:55:26', '2026-04-18 22:08:10'),
('43', 0, '合作伙伴', 1, '/list_43/', '', '', '', '', '', 255, 1, '2026-04-18 15:55:26', '2026-04-18 22:08:10'),
('44', 0, '行业培训', 1, '/list_44/', '', '', '', '', '', 255, 1, '2026-04-18 15:55:26', '2026-04-18 22:08:10'),
('45', 0, '人才队伍', 2, '/about_45/', '', '', '', '', '', 255, 1, '2026-04-18 15:57:05', '2026-04-18 22:08:10'),
('46', 0, '智库专家', 1, '/list_46/', '', '', '', '', '', 255, 1, '2026-04-18 15:57:53', '2026-04-18 22:08:10'),
('47', 0, '博士后创新实践基地', 1, '/list_47/', '', '', '', '', '', 255, 1, '2026-04-18 15:57:53', '2026-04-18 22:08:10'),
('48', 0, '博士后科研工作站', 1, '/list_48/', '', '', '', '', '', 255, 1, '2026-04-18 15:57:53', '2026-04-18 22:08:10'),
('49', 0, '人才招聘', 1, '/list_49/', '', '', '', '', '', 255, 1, '2026-04-18 15:57:53', '2026-04-18 22:08:10'),
('50', 0, '业务范畴', 2, '/about_50/', '', '', '', '', '', 254, 1, '2026-04-18 22:07:03', '2026-04-18 22:08:10'),
('51', 0, '互动交流', 2, '/about_51/', '', '', '', '', '', 255, 0, '2026-04-22 21:04:11', '2026-04-22 21:04:11'),
('52', 0, '在线咨询', 2, '/about_52/', '', '', '', '', '', 255, 0, '2026-04-22 21:04:45', '2026-04-22 21:06:42'),
('53', 0, '意见建议', 1, '/list_53/', '', '', '', '', '', 255, 0, '2026-04-22 21:05:24', '2026-04-22 21:05:24'),
('54', 0, '专家答疑', 1, '/list_54/', '', '', '', '', '', 255, 0, '2026-04-22 21:06:08', '2026-04-22 21:06:38')
ON CONFLICT ("code") DO NOTHING;

UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '3' AND parent."code" = '2' AND child."parent_id" = 0 AND child."updated_by" = 0;
UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '4' AND parent."code" = '2' AND child."parent_id" = 0 AND child."updated_by" = 0;
UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '12' AND parent."code" = '1' AND child."parent_id" = 0 AND child."updated_by" = 0;
UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '13' AND parent."code" = '1' AND child."parent_id" = 0 AND child."updated_by" = 0;
UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '14' AND parent."code" = '1' AND child."parent_id" = 0 AND child."updated_by" = 0;
UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '15' AND parent."code" = '45' AND child."parent_id" = 0 AND child."updated_by" = 0;
UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '16' AND parent."code" = '1' AND child."parent_id" = 0 AND child."updated_by" = 0;
UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '17' AND parent."code" = '4' AND child."parent_id" = 0 AND child."updated_by" = 0;
UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '18' AND parent."code" = '4' AND child."parent_id" = 0 AND child."updated_by" = 0;
UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '19' AND parent."code" = '5' AND child."parent_id" = 0 AND child."updated_by" = 0;
UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '20' AND parent."code" = '5' AND child."parent_id" = 0 AND child."updated_by" = 0;
UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '21' AND parent."code" = '5' AND child."parent_id" = 0 AND child."updated_by" = 0;
UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '22' AND parent."code" = '5' AND child."parent_id" = 0 AND child."updated_by" = 0;
UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '24' AND parent."code" = '23' AND child."parent_id" = 0 AND child."updated_by" = 0;
UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '25' AND parent."code" = '23' AND child."parent_id" = 0 AND child."updated_by" = 0;
UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '26' AND parent."code" = '25' AND child."parent_id" = 0 AND child."updated_by" = 0;
UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '27' AND parent."code" = '25' AND child."parent_id" = 0 AND child."updated_by" = 0;
UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '28' AND parent."code" = '25' AND child."parent_id" = 0 AND child."updated_by" = 0;
UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '29' AND parent."code" = '25' AND child."parent_id" = 0 AND child."updated_by" = 0;
UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '30' AND parent."code" = '25' AND child."parent_id" = 0 AND child."updated_by" = 0;
UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '32' AND parent."code" = '31' AND child."parent_id" = 0 AND child."updated_by" = 0;
UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '33' AND parent."code" = '31' AND child."parent_id" = 0 AND child."updated_by" = 0;
UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '35' AND parent."code" = '34' AND child."parent_id" = 0 AND child."updated_by" = 0;
UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '36' AND parent."code" = '34' AND child."parent_id" = 0 AND child."updated_by" = 0;
UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '37' AND parent."code" = '34' AND child."parent_id" = 0 AND child."updated_by" = 0;
UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '38' AND parent."code" = '34' AND child."parent_id" = 0 AND child."updated_by" = 0;
UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '39' AND parent."code" = '34' AND child."parent_id" = 0 AND child."updated_by" = 0;
UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '42' AND parent."code" = '40' AND child."parent_id" = 0 AND child."updated_by" = 0;
UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '43' AND parent."code" = '40' AND child."parent_id" = 0 AND child."updated_by" = 0;
UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '44' AND parent."code" = '40' AND child."parent_id" = 0 AND child."updated_by" = 0;
UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '46' AND parent."code" = '45' AND child."parent_id" = 0 AND child."updated_by" = 0;
UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '47' AND parent."code" = '45' AND child."parent_id" = 0 AND child."updated_by" = 0;
UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '48' AND parent."code" = '45' AND child."parent_id" = 0 AND child."updated_by" = 0;
UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '49' AND parent."code" = '45' AND child."parent_id" = 0 AND child."updated_by" = 0;
UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '50' AND parent."code" = '40' AND child."parent_id" = 0 AND child."updated_by" = 0;
UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '52' AND parent."code" = '51' AND child."parent_id" = 0 AND child."updated_by" = 0;
UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '53' AND parent."code" = '51' AND child."parent_id" = 0 AND child."updated_by" = 0;
UPDATE plugin_cms_category AS child SET "parent_id" = parent."id" FROM plugin_cms_category AS parent WHERE child."code" = '54' AND parent."code" = '51' AND child."parent_id" = 0 AND child."updated_by" = 0;

UPDATE plugin_cms_category
SET "list_template" = 'list.html', "content_template" = 'detail.html'
WHERE "type" = 1
  AND "updated_by" = 0
  AND BTRIM("list_template") = ''
  AND BTRIM("content_template") = '';

UPDATE plugin_cms_category
SET "list_template" = '', "content_template" = 'single.html'
WHERE "type" = 2
  AND "updated_by" = 0
  AND BTRIM("list_template") = ''
  AND BTRIM("content_template") = '';

UPDATE plugin_cms_category
SET "list_template" = 'list-card.html', "content_template" = 'detail.html'
WHERE "code" = '46'
  AND "updated_by" = 0
  AND "list_template" IN ('', 'list.html')
  AND "content_template" IN ('', 'detail.html');

INSERT INTO plugin_cms_slide ("group_code", "title", "subtitle", "image", "link", "sort", "status", "created_at", "updated_at")
SELECT seed."group_code", seed."title", seed."subtitle", seed."image", seed."link", seed."sort", seed."status", seed."created_at"::timestamp, seed."updated_at"::timestamp
FROM (VALUES
('1', '致力于先进材料前沿研究', '推动产业创新发展，打造国内一流科研平台', 'https://picsum.photos/seed/banner1/800/300', '', 255, 1, '2018-03-01 16:19:03', '2018-04-12 10:43:19'),
('1', '深化产学研合作', '联合高校科研机构，攻克关键核心技术', 'https://picsum.photos/seed/banner2/800/300', '', 255, 1, '2018-04-12 10:46:07', '2018-04-12 10:46:07'),
('1', '加速科技成果转化', '服务地方经济，推动先进材料产业升级', 'https://picsum.photos/seed/banner3/800/300', '', 255, 1, '2026-04-18 10:00:00', '2026-04-18 10:00:00')
) AS seed("group_code", "title", "subtitle", "image", "link", "sort", "status", "created_at", "updated_at")
WHERE NOT EXISTS (
    SELECT 1 FROM plugin_cms_slide AS existing
    WHERE existing."group_code" = seed."group_code"
      AND existing."title" = seed."title"
      AND existing."deleted_at" IS NULL
);

INSERT INTO plugin_cms_link ("group_code", "name", "url", "logo", "sort", "status", "created_at", "updated_at")
SELECT seed."group_code", seed."name", seed."url", seed."logo", seed."sort", seed."status", seed."created_at"::timestamp, seed."updated_at"::timestamp
FROM (VALUES
('1', '国家科技部', 'https://www.most.gov.cn', '', 255, 1, '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
('1', '江西省科技厅', 'http://kjt.jiangxi.gov.cn', '', 255, 1, '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
('1', '吉安市人民政府', 'http://www.jian.gov.cn', '', 255, 1, '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
('2', '江西日报', 'http://www.jxnews.com.cn', '', 255, 1, '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
('2', '科技日报', 'https://www.stdaily.com', '', 255, 1, '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
('2', '人民网', 'http://www.people.com.cn', '', 255, 1, '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
('3', '江西省科学院', 'http://www.jxas.ac.cn', '', 255, 1, '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
('3', '南昌大学', 'https://www.ncu.edu.cn', '', 255, 1, '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
('3', '东华大学', 'https://www.dhu.edu.cn', '', 255, 1, '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
('4', '中国材料网', 'http://www.matinfo.com.cn', '', 255, 1, '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
('4', '中国纺织网', 'http://www.texnet.com.cn', '', 255, 1, '2026-01-01 10:00:00', '2026-01-01 10:00:00')
) AS seed("group_code", "name", "url", "logo", "sort", "status", "created_at", "updated_at")
WHERE NOT EXISTS (
    SELECT 1 FROM plugin_cms_link AS existing
    WHERE existing."group_code" = seed."group_code"
      AND existing."name" = seed."name"
      AND existing."url" = seed."url"
      AND existing."deleted_at" IS NULL
);

INSERT INTO plugin_cms_article ("category_id", "title", "subtitle", "slug", "summary", "cover", "author", "source", "content", "tags", "keywords", "description", "sort", "status", "is_top", "is_recommend", "views", "published_at", "created_at", "updated_at") VALUES
((SELECT "id" FROM plugin_cms_category WHERE "code"='1'), '研究院简介', '', 'cms-1', '启明先进材料产业研究院聚焦先进材料公共技术服务、成果转化和产业协同。', '', 'admin', '本站', '<p>启明先进材料产业研究院面向先进材料产业链提供技术研发、检测验证、人才培训和成果转化服务，建设开放共享的创新服务平台。</p>', '', '', '启明先进材料产业研究院聚焦先进材料公共技术服务、成果转化和产业协同。', 255, 1, 0, 0, 61, '2018-04-11 17:26:11', '2018-04-11 17:26:11', '2026-04-18 21:57:09'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='4'), '产业智能化服务平台完成升级', '', 'cms-5', '研究院产业服务平台完成数据采集、项目协同和成果跟踪能力升级。', '', 'admin', '本站', '<p>研究院产业智能化服务平台完成升级，围绕项目申报、检测服务、成果转化和企业需求对接提供统一入口，提升服务响应效率。</p>', '', '', '研究院产业服务平台完成数据采集、项目协同和成果跟踪能力升级。', 255, 1, 0, 0, 4, '2018-04-12 09:52:36', '2018-04-12 10:06:15', '2018-04-13 09:36:44'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='4'), '青年科研交流活动顺利开展', '', 'cms-6', '青年科研交流活动围绕先进材料应用场景和项目协作展开。', '', 'admin', '本站', '<p>青年科研交流活动顺利开展，参会团队围绕先进材料应用场景、实验平台共建和企业需求对接进行了专题交流。</p>', '', '', '青年科研交流活动围绕先进材料应用场景和项目协作展开。', 255, 1, 0, 0, 5, '2018-04-12 10:06:22', '2018-04-12 10:08:03', '2018-04-13 09:36:25'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='4'), '青年科研交流与实验室安全培训完成年度首场宣讲', '', 'cms-7', '青年科研交流活动围绕先进材料应用场景、实验室安全和项目协作展开。', 'https://picsum.photos/seed/cms-cms-7/640/360', 'admin', '本站', '<p>研究院举办青年科研交流与实验室安全培训，围绕先进材料实验流程、样品流转、数据记录和安全操作展开宣讲，帮助青年科研人员更快熟悉公共平台规范。</p>', '文化建设,启明研究院,产业服务', '先进材料,文化建设,青年科研交流与实验室安全培训完成年度首场宣讲', '青年科研交流活动围绕先进材料应用场景、实验室安全和项目协作展开。', 255, 1, 0, 0, 40, '2018-04-12 10:08:50', '2018-04-12 10:09:37', '2026-04-18 21:58:25'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='12'), '建设背景', '', 'cms-18', '研究院建设背景围绕区域先进材料产业升级、公共技术服务和成果转化需求展开。', 'https://picsum.photos/seed/cms-cms-18/640/360', '超级管理员', '本站', '<p>启明先进材料产业研究院依托区域材料产业基础和公共服务需求建设，重点补齐企业在检测验证、中试放大、工艺优化和成果转化环节的公共能力短板。</p>', '建设背景,启明研究院,产业服务', '先进材料,建设背景,建设背景', '研究院建设背景围绕区域先进材料产业升级、公共技术服务和成果转化需求展开。', 255, 1, 0, 0, 34, '2026-04-16 16:56:52', '2026-04-16 16:56:52', '2026-04-18 21:59:20'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='13'), '建设背景', '', 'cms-19', '', '', '超级管理员', '本站', '', '', '', '', 255, 1, 0, 0, 16, '2026-04-18 15:35:49', '2026-04-18 15:35:49', '2026-04-18 15:35:49'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='14'), '组织机构', '', 'cms-20', '', '', '超级管理员', '本站', '', '', '', '', 255, 1, 0, 0, 14, '2026-04-18 15:37:31', '2026-04-18 15:37:31', '2026-04-18 15:37:31'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='16'), '联系我们', '', 'cms-21', '', '', '超级管理员', '本站', '', '', '', '', 255, 1, 0, 0, 15, '2026-04-18 15:41:11', '2026-04-18 15:41:11', '2026-04-18 15:41:11'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='17'), '院区风光', '', 'cms-22', '', 'https://picsum.photos/seed/honor1/120/90', '超级管理员', '本站', '', '', '', '', 255, 1, 0, 0, 16, '2026-04-18 15:45:00', '2026-04-18 15:45:00', '2026-04-19 09:40:02'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='18'), '品牌形象', '', 'cms-23', '', '', '超级管理员', '本站', '', '', '', '', 255, 1, 0, 0, 11, '2026-04-18 15:45:00', '2026-04-18 15:45:00', '2026-04-18 15:45:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='35'), '实验室规章制度', '', 'cms-24', '', '', '超级管理员', '本站', '', '', '', '', 255, 1, 0, 0, 9, '2026-04-18 15:51:02', '2026-04-18 15:51:02', '2026-04-18 15:51:02'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='36'), '设备名录', '', 'cms-25', '', '', '超级管理员', '本站', '', '', '', '', 255, 1, 0, 0, 11, '2026-04-18 15:51:02', '2026-04-18 15:51:02', '2026-04-18 15:51:02'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='37'), '预约指南与流程', '', 'cms-26', '', '', '超级管理员', '本站', '', '', '', '', 255, 1, 0, 0, 11, '2026-04-18 15:51:02', '2026-04-18 15:51:02', '2026-04-18 15:51:02'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='38'), '收费标准', '', 'cms-27', '', '', '超级管理员', '本站', '', '', '', '', 255, 1, 0, 0, 11, '2026-04-18 15:51:02', '2026-04-18 15:51:02', '2026-04-18 15:51:02'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='39'), '常见问题', '', 'cms-28', '', '', '超级管理员', '本站', '', '', '', '', 255, 1, 0, 0, 11, '2026-04-18 15:51:02', '2026-04-18 15:51:02', '2026-04-18 15:51:02'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='40'), '产业服务', '', 'cms-29', '', '', '超级管理员', '本站', '', '', '', '', 255, 1, 0, 0, 13, '2026-04-18 15:54:01', '2026-04-18 15:54:01', '2026-04-18 15:54:01'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='45'), '人才队伍', '', 'cms-30', '', '', '超级管理员', '本站', '', '', '', '', 255, 1, 0, 0, 11, '2026-04-18 15:57:05', '2026-04-18 15:57:05', '2026-04-18 15:57:05'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='19'), '研究院召开先进材料最新成果研讨会', '', 'cms-31', '', '', '', 'admin', '<p>3月21日研究院召开研讨会</p>', '', '', '', 255, 1, 0, 0, 10, '2026-03-21 10:00:00', '2026-03-21 10:00:00', '2026-03-21 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='19'), '专家团队赴企业调研科技服务需求', '', 'cms-32', '', '', '', 'admin', '<p>研究院组织专家团队赴企业调研</p>', '', '', '', 255, 1, 0, 0, 9, '2026-03-19 10:00:00', '2026-03-19 10:00:00', '2026-03-19 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='19'), '先进材料产业扶持政策解读讲座成功举办', '', 'cms-33', '', '', '', 'admin', '<p>研究院联合县工信局举办政策解读讲座</p>', '', '', '', 255, 1, 0, 0, 9, '2026-03-17 10:00:00', '2026-03-17 10:00:00', '2026-03-17 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='19'), '关于开展2026年度科研项目申报的通知', '', 'cms-34', '', '', '', 'admin', '<p>现启动2026年度科研项目申报工作</p>', '', '', '', 255, 1, 0, 0, 9, '2026-03-15 10:00:00', '2026-03-15 10:00:00', '2026-03-15 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='19'), '我院科研人员在国际顶级期刊发表论文', '', 'cms-35', '', '', '', 'admin', '<p>科研团队在高性能纤维复合材料领域取得重要突破</p>', '', '', '', 255, 1, 0, 0, 8, '2026-03-10 10:00:00', '2026-03-10 10:00:00', '2026-03-10 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='19'), '先进材料中试基地正式挂牌', '', 'cms-36', '', '', '', 'admin', '<p>中试基地正式挂牌运营</p>', '', '', '', 255, 1, 0, 0, 8, '2026-03-05 10:00:00', '2026-03-05 10:00:00', '2026-03-05 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='19'), '研究院与东华大学签署战略合作协议', '', 'cms-37', '', '', '', 'admin', '<p>双方签署战略合作协议</p>', '', '', '', 255, 1, 0, 0, 8, '2026-02-28 10:00:00', '2026-02-28 10:00:00', '2026-02-28 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='19'), '省科技厅领导莅临研究院指导工作', '', 'cms-38', '', '', '', 'admin', '<p>省科技厅领导莅临研究院考察指导</p>', '', '', '', 255, 1, 0, 0, 8, '2026-02-25 10:00:00', '2026-02-25 10:00:00', '2026-02-25 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='19'), '研究院2025年度工作总结暨表彰大会', '', 'cms-39', '', '', '', 'admin', '<p>研究院召开年度总结暨表彰大会</p>', '', '', '', 255, 1, 0, 0, 8, '2026-02-18 10:00:00', '2026-02-18 10:00:00', '2026-02-18 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='19'), '先进材料检测中心通过CMA资质认定', '', 'cms-40', '', '', '', 'admin', '<p>检测中心顺利通过CMA资质认定</p>', '', '', '', 255, 1, 0, 0, 8, '2026-02-10 10:00:00', '2026-02-10 10:00:00', '2026-02-10 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='21'), '市级新材料产业扶持政策及申报指南发布', '', 'cms-41', '', '', '', 'admin', '<p>市工信局联合发布扶持政策</p>', '', '', '', 255, 1, 0, 0, 8, '2026-03-20 10:00:00', '2026-03-20 10:00:00', '2026-03-20 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='21'), '国家高新技术企业认定管理办法修订版', '', 'cms-42', '', '', '', 'admin', '<p>科技部联合发布修订后的管理办法</p>', '', '', '', 255, 1, 0, 0, 8, '2026-03-12 10:00:00', '2026-03-12 10:00:00', '2026-03-12 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='21'), '江西省科技创新券申领及使用细则', '', 'cms-43', '', '', '', 'admin', '<p>省科技厅发布实施细则</p>', '', '', '', 255, 1, 0, 0, 8, '2026-03-01 10:00:00', '2026-03-01 10:00:00', '2026-03-01 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='21'), '促进科技成果转移转化的若干措施', '', 'cms-44', '', '', '', 'admin', '<p>国务院办公厅印发相关措施</p>', '', '', '', 255, 1, 0, 0, 8, '2026-02-25 10:00:00', '2026-02-25 10:00:00', '2026-02-25 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='21'), '知识产权资助及奖励办法最新公告', '', 'cms-45', '', '', '', 'admin', '<p>国家知识产权局发布最新奖励办法</p>', '', '', '', 255, 1, 0, 0, 8, '2026-02-18 10:00:00', '2026-02-18 10:00:00', '2026-02-18 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='21'), '高端人才引进配套服务保障政策', '', 'cms-46', '', '', '', 'admin', '<p>省人社厅发布人才引进政策</p>', '', '', '', 255, 1, 0, 0, 8, '2026-02-10 10:00:00', '2026-02-10 10:00:00', '2026-02-10 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='20'), '2026年清明节放假安排的通知', '', 'cms-47', '', '', '', 'admin', '<p>根据国务院办公厅通知精神发布放假安排</p>', '', '', '', 255, 1, 0, 0, 8, '2026-04-01 10:00:00', '2026-04-01 10:00:00', '2026-04-01 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='20'), '开展实验室安全大检查的通知', '', 'cms-48', '', '', '', 'admin', '<p>为确保实验室安全运行开展大检查</p>', '', '', '', 255, 1, 0, 0, 9, '2026-03-28 10:00:00', '2026-03-28 10:00:00', '2026-03-28 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='20'), '2026年度研究院公开招聘公告', '', 'cms-49', '', '', '', 'admin', '<p>面向社会公开招聘科研人员3名</p>', '', '', '', 255, 1, 0, 0, 8, '2026-03-15 10:00:00', '2026-03-15 10:00:00', '2026-03-15 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='22'), '区域打造先进材料产业高地', '', 'cms-50', '', '', '', 'admin', '<p>江西日报报道示范区推动产业转型升级</p>', '', '', '', 255, 1, 0, 0, 2, '2026-03-20 10:00:00', '2026-03-20 10:00:00', '2026-03-20 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='22'), '产学研深度融合助力先进材料产业发展', '', 'cms-51', '', '', '', 'admin', '<p>科技日报专题报道研究院产学研合作模式</p>', '', '', '', 255, 1, 0, 0, 8, '2026-03-10 10:00:00', '2026-03-10 10:00:00', '2026-03-10 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='25'), '国家级先进材料技术学术报告发布国家级先进材料技术学术报告发布国家级先进材料技术学术报告发布国家级先进材料技术学术报告发布', '', 'cms-52', '发布高性能纤维复合材料技术发展白皮书', '', '', 'admin', '&lt;p&gt;发布高性能纤维复合材料技术发展白皮书&lt;/p&gt;', '', '', '发布高性能纤维复合材料技术发展白皮书', 255, 1, 0, 0, 12, '2026-03-20 10:00:00', '2026-03-20 10:00:00', '2026-04-18 17:37:07'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='25'), '新型农业材料技术转化成功', '', 'cms-53', '', '', '', 'admin', '<p>新型农业覆盖材料技术实现产业化</p>', '', '', '', 255, 1, 0, 0, 5, '2026-03-15 10:00:00', '2026-03-15 10:00:00', '2026-03-15 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='25'), '攻克无纺布生产关键技术瓶颈', '', 'cms-54', '', '', '', 'admin', '<p>产品强度提升30%以上</p>', '', '', '', 255, 1, 0, 0, 10, '2026-03-08 10:00:00', '2026-03-08 10:00:00', '2026-03-08 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='25'), '高纯度纳米材料制备取得进展', '', 'cms-55', '', '', '', 'admin', '<p>纯度可达99.99%</p>', '', '', '', 255, 1, 0, 0, 11, '2026-02-28 10:00:00', '2026-02-28 10:00:00', '2026-02-28 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='25'), '高性能纤维复合膜应用研究通过验收', '', 'cms-56', '', '', '', 'admin', '<p>海水淡化应用项目通过省级验收</p>', '', '', '', 255, 1, 0, 0, 8, '2026-02-20 10:00:00', '2026-02-20 10:00:00', '2026-02-20 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='43'), '江西省科学院产学研合作基地', '', 'cms-57', '', '', '', 'admin', '<p>共建产学研合作基地</p>', '', '', '', 255, 1, 0, 0, 8, '2026-01-10 10:00:00', '2026-01-10 10:00:00', '2026-01-10 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='43'), '启明先进材料科技有限公司', '', 'cms-58', '', '', '', 'admin', '<p>核心合作企业</p>', '', '', '', 255, 1, 0, 0, 8, '2026-01-10 10:00:00', '2026-01-10 10:00:00', '2026-01-10 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='43'), '东华大学材料科学与工程学院', '', 'cms-59', '', '', '', 'admin', '<p>纤维材料领域深度合作</p>', '', '', '', 255, 1, 0, 0, 8, '2026-01-10 10:00:00', '2026-01-10 10:00:00', '2026-01-10 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='43'), '长三角新材料产业创新联盟', '', 'cms-60', '', '', '', 'admin', '<p>拓展产业合作网络</p>', '', '', '', 255, 1, 0, 0, 9, '2026-01-10 10:00:00', '2026-01-10 10:00:00', '2026-01-10 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='43'), '省内重点企业协同创新联合体', '', 'cms-61', '', '', '', 'admin', '<p>牵头组建协同创新联合体</p>', '', '', '', 255, 1, 0, 0, 10, '2026-01-10 10:00:00', '2026-01-10 10:00:00', '2026-01-10 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='32'), '高强度功能合成革生产技术授权公告', '', 'cms-62', '', '', '', 'admin', '<p>技术获专利授权</p>', '', '', '', 255, 1, 0, 0, 8, '2026-03-15 10:00:00', '2026-03-15 10:00:00', '2026-03-15 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='32'), '年产万吨环保型材料项目签约', '', 'cms-63', '', '', '', 'admin', '<p>与工业园区签署合作协议</p>', '', '', '', 255, 1, 0, 0, 10, '2026-03-08 10:00:00', '2026-03-08 10:00:00', '2026-03-08 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='32'), '水性聚氨酯树脂制备专利转让', '', 'cms-64', '', '', '', 'admin', '<p>专利成功转让</p>', '', '', '', 255, 1, 0, 0, 10, '2026-02-28 10:00:00', '2026-02-28 10:00:00', '2026-02-28 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='32'), '新型涂层工艺完成企业技术对接', '', 'cms-65', '', '', '', 'admin', '<p>环保涂层工艺完成对接</p>', '', '', '', 255, 1, 0, 0, 8, '2026-02-20 10:00:00', '2026-02-20 10:00:00', '2026-02-20 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='32'), '高性能复合材料中试技术完成移交', '', 'cms-66', '', '', '', 'admin', '<p>中试技术完成技术移交</p>', '', '', '', 255, 1, 0, 0, 8, '2026-02-10 10:00:00', '2026-02-10 10:00:00', '2026-02-10 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='33'), '高端透气膜项目进入量产阶段', '', 'cms-67', '', '', '', 'admin', '<p>技术成果实现产业化</p>', '', '', '', 255, 1, 0, 0, 8, '2026-03-15 10:00:00', '2026-03-15 10:00:00', '2026-03-15 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='33'), '汽车内饰用先进材料通过国际认证', '', 'cms-68', '', '', '', 'admin', '<p>通过国际汽车行业质量体系认证</p>', '', '', '', 255, 1, 0, 0, 11, '2026-03-05 10:00:00', '2026-03-05 10:00:00', '2026-03-05 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='33'), '废旧布料循环利用技术落地成效', '', 'cms-69', '', '', '', 'admin', '<p>年处理废旧布料达5000吨</p>', '', '', '', 255, 1, 0, 0, 9, '2026-02-25 10:00:00', '2026-02-25 10:00:00', '2026-02-25 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='33'), '环保先进材料项目实现批量出货', '', 'cms-70', '', '', '', 'admin', '<p>月产量达到50万米</p>', '', '', '', 255, 1, 0, 0, 7, '2026-02-15 10:00:00', '2026-02-15 10:00:00', '2026-02-15 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='33'), '功能性新材料示范线投产运行稳定', '', 'cms-71', '', '', '', 'admin', '<p>产品质量达到预期目标</p>', '', '', '', 255, 1, 0, 0, 8, '2026-02-05 10:00:00', '2026-02-05 10:00:00', '2026-02-05 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='46'), '王教授', '', 'cms-72', '新材料研究专家', 'https://picsum.photos/seed/expert1/80/105', '', 'admin', '<p>东华大学材料学院教授</p>', '', '', '新材料研究专家', 255, 1, 0, 0, 8, '2026-01-01 10:00:00', '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='46'), '李博士', '', 'cms-73', '高分子材料专家', 'https://picsum.photos/seed/expert2/80/105', '', 'admin', '<p>中科院化学研究所博士</p>', '', '', '高分子材料专家', 255, 1, 0, 0, 8, '2026-01-01 10:00:00', '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='46'), '陈教授', '', 'cms-74', '纳米材料专家', 'https://picsum.photos/seed/expert3/80/105', '', 'admin', '<p>南昌大学化学学院教授</p>', '', '', '纳米材料专家', 255, 1, 0, 0, 8, '2026-01-01 10:00:00', '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='46'), '张主任', '', 'cms-75', '纺织工艺专家', 'https://picsum.photos/seed/expert4/80/105', '', 'admin', '<p>省先进材料工程技术研究中心主任</p>', '', '', '纺织工艺专家', 255, 1, 0, 0, 8, '2026-01-01 10:00:00', '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='46'), '刘教授', '', 'cms-76', '复合材料专家', 'https://picsum.photos/seed/expert5/80/105', '', 'admin', '<p>华东理工大学材料学院教授</p>', '', '', '复合材料专家', 255, 1, 0, 0, 8, '2026-01-01 10:00:00', '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='46'), '赵工', '', 'cms-77', '工艺流程专家', 'https://picsum.photos/seed/expert6/80/105', '', 'admin', '<p>启明先进材料科技公司总工程师</p>', '', '', '工艺流程专家', 255, 1, 0, 0, 9, '2026-01-01 10:00:00', '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='46'), '孙教授', '', 'cms-78', '纺织工程专家', 'https://picsum.photos/seed/expert7/80/105', '', 'admin', '<p>武汉纺织大学纺织学院教授</p>', '', '', '纺织工程专家', 255, 1, 0, 0, 8, '2026-01-01 10:00:00', '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='46'), '周博士', '', 'cms-79', '化学工程专家', 'https://picsum.photos/seed/expert8/80/105', '', 'admin', '<p>清华大学化学工程系博士</p>', '', '', '化学工程专家', 255, 1, 0, 0, 9, '2026-01-01 10:00:00', '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='46'), '吴研究员', '', 'cms-80', '材料分析专家', 'https://picsum.photos/seed/expert9/80/105', '', 'admin', '<p>中科院上海硅酸盐研究所研究员</p>', '', '', '材料分析专家', 255, 1, 0, 0, 10, '2026-01-01 10:00:00', '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='46'), '郑教授', '', 'cms-81', '环境材料专家', 'https://picsum.photos/seed/expert10/80/105', '', 'admin', '<p>北京化工大学教授</p>', '', '', '环境材料专家', 255, 1, 0, 0, 10, '2026-01-01 10:00:00', '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='46'), '黄主任', '', 'cms-82', '产业化专家', 'https://picsum.photos/seed/expert11/80/105', '', 'admin', '<p>区域科技协同服务办公室主任</p>', '', '', '产业化专家', 255, 1, 0, 0, 11, '2026-01-01 10:00:00', '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='46'), '林博士', '', 'cms-83', '复合材料专家', 'https://picsum.photos/seed/expert12/80/105', '', 'admin', '<p>北京化工大学博士</p>', '', '', '复合材料专家', 255, 1, 0, 0, 12, '2026-01-01 10:00:00', '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='17'), '国家高新技术企业', '', 'cms-84', '', 'https://picsum.photos/seed/honor1/120/90', '', 'admin', '<p>国家高新技术企业</p>', '', '', '', 255, 1, 0, 0, 12, '2026-01-01 10:00:00', '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='17'), '省级重点实验室', '', 'cms-85', '', 'https://picsum.photos/seed/honor2/120/90', '', 'admin', '<p>省级重点实验室</p>', '', '', '', 255, 1, 0, 0, 8, '2026-01-01 10:00:00', '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='17'), '科技进步奖', '', 'cms-86', '', 'https://picsum.photos/seed/honor3/120/90', '', 'admin', '<p>科技进步奖</p>', '', '', '', 255, 1, 0, 0, 9, '2026-01-01 10:00:00', '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='17'), '产学研合作基地', '', 'cms-87', '', 'https://picsum.photos/seed/honor4/120/90', '', 'admin', '<p>产学研合作基地</p>', '', '', '', 255, 1, 0, 0, 9, '2026-01-01 10:00:00', '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='17'), '先进集体', '', 'cms-88', '', 'https://picsum.photos/seed/honor5/120/90', '', 'admin', '<p>先进集体</p>', '', '', '', 255, 1, 0, 0, 9, '2026-01-01 10:00:00', '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='17'), '质量管理认证', '', 'cms-89', '', 'https://picsum.photos/seed/honor6/120/90', '', 'admin', '<p>质量管理认证</p>', '', '', '', 255, 1, 0, 0, 10, '2026-01-01 10:00:00', '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='17'), '先进荣誉单位', '', 'cms-90', '', 'https://picsum.photos/seed/honor7/120/90', '', 'admin', '<p>先进荣誉单位</p>', '', '', '', 255, 1, 0, 0, 10, '2026-01-01 10:00:00', '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='17'), '技术创新中心', '', 'cms-91', '', 'https://picsum.photos/seed/honor8/120/90', '', 'admin', '<p>技术创新中心</p>', '', '', '', 255, 1, 0, 0, 9, '2026-01-01 10:00:00', '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='17'), '知识产权示范', '', 'cms-92', '', 'https://picsum.photos/seed/honor9/120/90', '', 'admin', '<p>知识产权示范</p>', '', '', '', 255, 1, 0, 0, 8, '2026-01-01 10:00:00', '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='17'), '优秀科研团队', '', 'cms-93', '', 'https://picsum.photos/seed/honor10/120/90', '', 'admin', '<p>优秀科研团队</p>', '', '', '', 255, 1, 0, 0, 13, '2026-01-01 10:00:00', '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='36'), '国家重点实验室仪器预约开放', '', 'cms-94', '', '', '', 'admin', '<p>多台高精度分析仪器面向社会开放预约使用</p>', '', '', '', 255, 1, 0, 0, 8, '2026-03-20 10:00:00', '2026-03-20 10:00:00', '2026-03-20 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='36'), '高倍电子显微镜共享使用说明', '', 'cms-95', '', '', '', 'admin', '<p>配备日立SU8010冷场发射扫描电子显微镜</p>', '', '', '', 255, 1, 0, 0, 8, '2026-03-15 10:00:00', '2026-03-15 10:00:00', '2026-03-15 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='36'), '新引进热分析光谱仪投入使用', '', 'cms-96', '', '', '', 'admin', '<p>新引进的同步热分析仪STA449F3已正式投入使用</p>', '', '', '', 255, 1, 0, 0, 14, '2026-03-08 10:00:00', '2026-03-08 10:00:00', '2026-03-08 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='36'), '公共检测平台设备名录更新', '', 'cms-97', '', '', '', 'admin', '<p>更新包括万能材料试验机、纤维细度分析仪等设备</p>', '', '', '', 255, 1, 0, 0, 8, '2026-02-28 10:00:00', '2026-02-28 10:00:00', '2026-02-28 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='36'), '公共检测平台本月开放时段公布', '', 'cms-98', '', '', '', 'admin', '<p>3月检测平台开放时段已公布，欢迎预约</p>', '', '', '', 255, 1, 0, 0, 7, '2026-03-01 10:00:00', '2026-03-01 10:00:00', '2026-03-01 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='46'), '何教授', '', 'cms-99', '功能材料专家', 'https://picsum.photos/seed/expert13/80/105', '', 'admin', '<p>浙江大学材料学院教授，功能材料方向</p>', '', '', '功能材料专家', 255, 1, 0, 0, 8, '2026-01-01 10:00:00', '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='46'), '杨博士', '', 'cms-100', '分析化学专家', 'https://picsum.photos/seed/expert14/80/105', '', 'admin', '<p>中科院福建物质结构研究所博士</p>', '', '', '分析化学专家', 255, 1, 0, 0, 8, '2026-01-01 10:00:00', '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='46'), '朱研究员', '', 'cms-101', '纺织材料专家', 'https://picsum.photos/seed/expert15/80/105', '', 'admin', '<p>东华大学纺织研究院研究员</p>', '', '', '纺织材料专家', 255, 1, 0, 0, 8, '2026-01-01 10:00:00', '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='46'), '徐教授', '', 'cms-102', '高分子专家', 'https://picsum.photos/seed/expert16/80/105', '', 'admin', '<p>四川大学高分子学院教授</p>', '', '', '高分子专家', 255, 1, 0, 0, 5, '2026-01-01 10:00:00', '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='46'), '马主任', '', 'cms-103', '质量检测专家', 'https://picsum.photos/seed/expert17/80/105', '', 'admin', '<p>国家纺织品质检中心高级工程师</p>', '', '', '质量检测专家', 255, 1, 0, 0, 8, '2026-01-01 10:00:00', '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='46'), '胡博士', '', 'cms-104', '纳米纤维专家', 'https://picsum.photos/seed/expert18/80/105', '', 'admin', '<p>苏州大学纺织与服装工程学院博士</p>', '', '', '纳米纤维专家', 255, 1, 0, 0, 8, '2026-01-01 10:00:00', '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='46'), '郭教授', '', 'cms-105', '染色技术专家', 'https://picsum.photos/seed/expert19/80/105', '', 'admin', '<p>天津工业大学纺织学院教授</p>', '', '', '染色技术专家', 255, 1, 0, 0, 8, '2026-01-01 10:00:00', '2026-01-01 10:00:00', '2026-01-01 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='46'), '罗研究员', '', 'cms-106', '材料力学专家', 'https://picsum.photos/seed/expert20/80/105', '', 'admin', '<p>罗研究员长期从事材料力学性能评价与失效分析，可为企业提供样品测试方案设计、数据解读和工艺改进建议。</p>', '', '', '材料力学专家', 255, 1, 0, 0, 10, '2026-01-01 10:00:00', '2026-01-01 10:00:00', '2026-04-18 21:52:32'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='25'), '新型抗菌先进材料研发取得突破', '', 'cms-107', '', '', '', 'admin', '<p>研究院成功研发具有持久抗菌功能的先进材料</p>', '', '', '', 255, 1, 0, 0, 8, '2026-02-15 10:00:00', '2026-02-15 10:00:00', '2026-02-15 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='25'), '高性能复合纤维项目通过省级鉴定', '', 'cms-108', '', '', '', 'admin', '<p>高性能复合纤维产业化关键技术项目通过省级科技成果鉴定</p>', '', '', '', 255, 1, 0, 0, 8, '2026-02-08 10:00:00', '2026-02-08 10:00:00', '2026-02-08 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='25'), '环保型水基先进材料生产技术开发成功', '', 'cms-109', '', '', '', 'admin', '<p>完全替代有机溶剂的环保型水基先进材料生产技术</p>', '', '', '', 255, 1, 0, 0, 8, '2026-01-25 10:00:00', '2026-01-25 10:00:00', '2026-01-25 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='25'), '先进材料在医疗领域应用研究进展', '', 'cms-110', '', '', '', 'admin', '<p>先进材料在医用敷料和手术缝合线领域的应用取得新进展</p>', '', '', '', 255, 1, 0, 0, 8, '2026-01-18 10:00:00', '2026-01-18 10:00:00', '2026-01-18 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='25'), '智能响应纤维材料实验室成果发表', '', 'cms-111', '', '', '', 'admin', '<p>温敏和pH响应型智能纤维材料研究成果发表于国际期刊</p>', '', '', '', 255, 1, 0, 0, 8, '2026-01-10 10:00:00', '2026-01-10 10:00:00', '2026-01-10 10:00:00'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='50'), '业务范畴', '', 'cms-113', '', '', '超级管理员', '本站', '', '', '', '', 255, 1, 0, 0, 11, '2026-04-18 22:07:03', '2026-04-18 22:07:03', '2026-04-18 22:07:03'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='51'), '互动交流', '', 'cms-114', '', '', '超级管理员', '本站', '', '', '', '', 255, 1, 0, 0, 8, '2026-04-22 21:04:11', '2026-04-22 21:04:11', '2026-04-22 21:04:11'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='52'), '在线咨询服务指南', '', 'cms-115', '在线咨询窗口、工作时间、联系电话和服务材料说明。', 'https://picsum.photos/seed/cms-cms-115/640/360', '超级管理员', '本站', '<p>在线咨询服务面向企业技术负责人、项目申报人员和高校课题组开放，访客可提交材料方向、样品状态、预期指标、时间要求和联系方式，研究院工作人员会在工作日内完成初步分流。</p>', '在线咨询,启明研究院,产业服务', '先进材料,在线咨询,服务指南', '在线咨询窗口、工作时间、联系电话和服务材料说明。', 255, 1, 0, 0, 15, '2026-04-22 21:04:45', '2026-04-22 21:04:45', '2026-04-22 21:07:26'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='54'), '常见技术问题解答', '', 'cms-116', '整理企业在材料检测、样品准备、报告交付和成果转化咨询中的高频问题。', 'https://picsum.photos/seed/cms-cms-116/640/360', '超级管理员', '本站', '<p>本栏目整理企业和课题组在检测预约、样品数量、测试周期、报告格式、专家咨询和后续转化服务中的常见问题，便于访客在提交咨询前了解基础规则。</p>', '常见问题,启明研究院,产业服务', '先进材料,常见技术问题解答,咨询服务', '整理企业在材料检测、样品准备、报告交付和成果转化咨询中的高频问题。', 255, 1, 0, 0, 3, '2026-04-22 21:07:49', '2026-04-22 21:08:09', '2026-04-22 21:08:09'),
((SELECT "id" FROM plugin_cms_category WHERE "code"='54'), '专家团队与服务方向介绍', '', 'cms-117', '介绍研究院专家团队的材料设计、工艺放大、检测评价和产业化服务方向。', 'https://picsum.photos/seed/cms-cms-117/640/360', '超级管理员', '本站', '<p>专家团队覆盖高分子材料、纤维复合材料、涂层工艺、性能检测、质量体系和成果转化等方向，可为企业提供需求诊断、方案评审和项目跟踪服务。</p>', '专家团队,启明研究院,产业服务', '先进材料,专家团队,咨询服务', '介绍研究院专家团队的材料设计、工艺放大、检测评价和产业化服务方向。', 255, 1, 0, 0, 7, '2026-04-22 21:08:10', '2026-04-22 21:08:27', '2026-04-22 21:08:27')
ON CONFLICT ("slug") DO NOTHING;

UPDATE plugin_cms_article
SET "title" = CASE "slug"
    WHEN 'cms-7' THEN '青年科研交流与实验室安全培训完成年度首场宣讲'
    WHEN 'cms-19' THEN '研究院概况'
    WHEN 'cms-52' THEN '高性能纤维复合材料技术白皮书发布'
    WHEN 'cms-115' THEN '在线咨询服务指南'
    WHEN 'cms-116' THEN '常见技术问题解答'
    WHEN 'cms-117' THEN '专家团队与服务方向介绍'
    ELSE "title"
END
WHERE "updated_by" = 0 AND "slug" IN ('cms-7', 'cms-19', 'cms-52', 'cms-115', 'cms-116', 'cms-117');

UPDATE plugin_cms_article AS article
SET
    "summary" = CASE
        WHEN category."code" = '46' THEN CASE article."slug"
            WHEN 'cms-72' THEN '教授'
            WHEN 'cms-73' THEN '博士'
            WHEN 'cms-74' THEN '教授'
            WHEN 'cms-75' THEN '主任'
            WHEN 'cms-76' THEN '教授'
            WHEN 'cms-77' THEN '总工程师'
            WHEN 'cms-78' THEN '教授'
            WHEN 'cms-79' THEN '博士'
            WHEN 'cms-80' THEN '研究员'
            WHEN 'cms-81' THEN '教授'
            WHEN 'cms-82' THEN '主任'
            WHEN 'cms-83' THEN '博士'
            WHEN 'cms-99' THEN '教授'
            WHEN 'cms-100' THEN '博士'
            WHEN 'cms-101' THEN '研究员'
            WHEN 'cms-102' THEN '副教授'
            WHEN 'cms-103' THEN '高工'
            WHEN 'cms-104' THEN '博士'
            WHEN 'cms-105' THEN '教授'
            WHEN 'cms-106' THEN '超纤专家'
            ELSE LEFT(REGEXP_REPLACE(article."description", '(专家|研究员)$', ''), 6)
        END
        ELSE LEFT(CONCAT('《', article."title", '》围绕', category."name", '场景补充背景、流程、服务触点和阶段成效，适合用于启明先进材料产业研究院官网演示。'), 260)
    END,
    "description" = LEFT(CONCAT('启明先进材料产业研究院在', category."name", '栏目发布《', article."title", '》，展示先进材料研发、检测验证、成果转化和公共服务能力。'), 300),
    "cover" = CASE
        WHEN BTRIM(article."cover") = '' THEN CONCAT('https://picsum.photos/seed/cms-', article."slug", '/640/360')
        ELSE article."cover"
    END,
    "author" = CASE
        WHEN BTRIM(article."author") = '' THEN '本站编辑'
        ELSE article."author"
    END,
    "source" = CASE
        WHEN BTRIM(article."source") = '' THEN '本站'
        ELSE article."source"
    END,
    "keywords" = LEFT(CONCAT('先进材料,', category."name", ',', article."title"), 255),
    "tags" = LEFT(CONCAT(category."name", ',启明研究院,产业服务'), 255),
    "content" = CONCAT(
        '<p>《', article."title", '》是启明先进材料产业研究院在“', category."name", '”栏目中的演示内容，面向企业技术负责人、高校课题组、园区服务人员和来访客户说明相关工作。文章以先进材料研发、检测验证、成果转化和公共服务为主线，补充背景、流程、成效和后续安排，让访客在演示站点中看到完整而可信的业务表达。</p>',
        CASE
            WHEN category."code" IN ('1', '12', '13', '14', '16') THEN '<p>该内容重点呈现研究院的建设定位、组织协同、服务窗口和运营机制。研究院围绕高性能纤维、绿色涂层、功能膜材料、循环利用和智能检测方向，连接科研团队、企业工厂和产业园区，形成从需求诊断、方案评审、中试验证到交付跟踪的闭环。</p>'
            WHEN category."code" IN ('35', '36', '37', '38', '39') THEN '<p>公共平台以样品接收、实验排期、检测记录、报告复核和设备维护为核心流程，开放电子显微、热分析、力学测试、纤维表征和环境可靠性等能力。企业可按项目阶段提交样品和指标要求，平台工程师会给出预约建议、测试边界和数据交付说明。</p>'
            WHEN category."code" IN ('19', '20', '21', '22', '25', '31', '32', '33') THEN '<p>本条资讯按照官网发布口径整理，补充了活动背景、参与主体、关键议题和落地价值。内容覆盖政策申报、项目合作、科研进展、平台开放和成果应用，便于访客理解研究院如何把技术资源转化为企业可使用的服务。</p>'
            WHEN category."code" IN ('40', '50', '51', '52', '54') THEN '<p>服务内容围绕企业咨询、项目诊断、专家评审、政策辅导和后续跟踪展开，强调可预约、可对接、可反馈的线上线下联动方式。访客可以据此了解业务入口、所需材料、办理节奏和常见注意事项，演示时不再只看到空白栏目。</p>'
            WHEN category."code" IN ('43') THEN '<p>合作内容聚焦产学研协同、联合攻关、人才培养和成果落地，展示研究院与高校院所、产业联盟、园区企业之间的常态化连接。通过共建课题、共享平台和共同服务企业，合作方可以把实验室成果更快转化为中试验证和示范应用。</p>'
            WHEN category."code" IN ('45', '46') THEN '<p>人才内容突出专家团队的专业方向、服务能力和项目经验，覆盖材料设计、工艺放大、质量检测、知识产权和产业化管理等环节。演示站点通过这些信息帮助访客判断适合对接的专家资源，也能呈现研究院的人才厚度和协同能力。</p>'
            WHEN category."code" IN ('17', '18') THEN '<p>展示内容用于呈现院区环境、品牌形象、荣誉成果和开放场景，配合图片让官网首页、列表页和详情页更有现场感。访客可以从中了解研究院的公共平台基础、服务氛围、合作资质和面向产业的长期投入。</p>'
            ELSE '<p>该栏目围绕研究院日常运营和产业服务整理内容，强调信息完整、口径清晰、入口明确。演示数据覆盖背景说明、资源能力、服务流程和联系人线索，方便产品演示时从首页跳转到详情页后仍能看到充实内容。</p>'
        END,
        '<p>在业务流程上，研究院通常先收集企业或课题组的技术需求，组织材料、工艺、检测和产业化专家进行初步研判，再根据样品状态、指标目标和交付周期制定服务方案。需要进入中试或检测环节的项目，会同步记录样品编号、设备安排、数据报告和风险提示，确保后续复盘有据可查。</p>',
        '<p>这批初始化内容主要用于 CMS 插件演示和二次开发参考，正文均保持三百字以上，并尽量贴近真实官网的表达方式。后续用户可以在后台继续替换封面、调整摘要、补充附件或关联留言反馈，把演示站点扩展成正式的机构官网、公共服务平台或产业园门户。</p>'
    )
FROM plugin_cms_category AS category
WHERE article."category_id" = category."id"
  AND article."slug" IN ('cms-1', 'cms-5', 'cms-6', 'cms-7', 'cms-18', 'cms-19', 'cms-20', 'cms-21', 'cms-22', 'cms-23', 'cms-24', 'cms-25', 'cms-26', 'cms-27', 'cms-28', 'cms-29', 'cms-30', 'cms-31', 'cms-32', 'cms-33', 'cms-34', 'cms-35', 'cms-36', 'cms-37', 'cms-38', 'cms-39', 'cms-40', 'cms-41', 'cms-42', 'cms-43', 'cms-44', 'cms-45', 'cms-46', 'cms-47', 'cms-48', 'cms-49', 'cms-50', 'cms-51', 'cms-52', 'cms-53', 'cms-54', 'cms-55', 'cms-56', 'cms-57', 'cms-58', 'cms-59', 'cms-60', 'cms-61', 'cms-62', 'cms-63', 'cms-64', 'cms-65', 'cms-66', 'cms-67', 'cms-68', 'cms-69', 'cms-70', 'cms-71', 'cms-72', 'cms-73', 'cms-74', 'cms-75', 'cms-76', 'cms-77', 'cms-78', 'cms-79', 'cms-80', 'cms-81', 'cms-82', 'cms-83', 'cms-84', 'cms-85', 'cms-86', 'cms-87', 'cms-88', 'cms-89', 'cms-90', 'cms-91', 'cms-92', 'cms-93', 'cms-94', 'cms-95', 'cms-96', 'cms-97', 'cms-98', 'cms-99', 'cms-100', 'cms-101', 'cms-102', 'cms-103', 'cms-104', 'cms-105', 'cms-106', 'cms-107', 'cms-108', 'cms-109', 'cms-110', 'cms-111', 'cms-113', 'cms-114', 'cms-115', 'cms-116', 'cms-117')
  AND article."updated_by" = 0
  AND (
      BTRIM(article."summary") = ''
      OR BTRIM(article."description") = ''
      OR BTRIM(article."cover") = ''
      OR BTRIM(article."content") = ''
      OR article."content" LIKE '&lt;%'
      OR LENGTH(REGEXP_REPLACE(article."content", '<[^>]+>', '', 'g')) < 300
      OR category."code" = '46'
  );

INSERT INTO plugin_cms_message ("name", "mobile", "email", "content", "reply", "status", "user_ip", "user_agent", "created_at", "updated_at")
SELECT seed."name", seed."mobile", seed."email", seed."content", seed."reply", seed."status", seed."user_ip", seed."user_agent", seed."created_at"::timestamp, seed."updated_at"::timestamp
FROM (VALUES
('吉安材料科技有限公司 王经理', '13888886666', 'wang@example.com', '想了解高性能复合材料检测服务的预约流程。', '您好，可以在工作日联系公共服务中心，我们会安排检测工程师对接样品和指标要求。', 1, '127.0.0.1', 'starter', '2026-05-09 14:00:00', '2026-05-09 14:20:00'),
('启明新材项目组', '13900001111', 'project@example.com', '希望咨询科技成果转化合作的入驻条件。', '您好，成果转化合作可先提交项目简介，研究院会组织专家进行初审和需求匹配。', 1, '127.0.0.1', 'starter', '2026-05-09 15:10:00', '2026-05-09 15:35:00'),
('高校联合实验室 李老师', '13700002222', 'li@example.edu', '共享平台设备是否支持校企联合课题预约？', '支持，请提前准备课题说明、样品信息和预计使用时段，工作人员会协助完成预约。', 1, '127.0.0.1', 'starter', '2026-05-09 16:00:00', '2026-05-09 16:30:00'),
('产业园企业代表', '13600003333', 'park@example.com', '想报名参加近期先进材料产业政策培训。', '', 0, '127.0.0.1', 'starter', '2026-05-09 17:00:00', '2026-05-09 17:00:00'),
('访客陈先生', '13500004444', 'chen@example.com', '请问专家答疑栏目能否提交更详细的技术问题？', '', 2, '127.0.0.1', 'starter', '2026-05-09 18:00:00', '2026-05-09 18:10:00')
) AS seed("name", "mobile", "email", "content", "reply", "status", "user_ip", "user_agent", "created_at", "updated_at")
WHERE NOT EXISTS (
    SELECT 1 FROM plugin_cms_message AS existing
    WHERE existing."name" = seed."name"
      AND existing."content" = seed."content"
      AND existing."deleted_at" IS NULL
);
