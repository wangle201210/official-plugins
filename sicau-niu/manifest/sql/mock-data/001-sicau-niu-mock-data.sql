-- ------------------------------------------------------------
-- sicau-niu mock/demo data
-- sicau-niu 模拟/演示数据
-- Purpose: populate every operator page and the public H5 wall with representative
--          records so the activity UI is not empty in demos.
-- Dialect: PostgreSQL. Idempotent via INSERT ... SELECT ... WHERE NOT EXISTS on the
--          real business keys; no explicit auto-increment id is written and all
--          foreign keys are resolved by business key, never hardcoded ids.
-- Data classification: demo/non-sensitive (synthetic openids/phones/nicknames).
-- ------------------------------------------------------------

-- 1) 院系字典 (college) — keyed by name.
INSERT INTO plugin_sicau_niu_college ("name")
SELECT v.name FROM (VALUES
    ('信息工程学院'), ('水利水电学院'), ('农学院'), ('动物科技学院'),
    ('风景园林学院'), ('食品学院'), ('经济学院'), ('林学院')
) AS v(name)
WHERE NOT EXISTS (SELECT 1 FROM plugin_sicau_niu_college c WHERE c."name" = v.name AND c."deleted_at" IS NULL);

-- 2) 牛 (niu) — keyed by code; college resolved by name.
INSERT INTO plugin_sicau_niu_niu ("code","niu_type","special_subtype","name","college_id","lat","lng","status","online_at")
SELECT v.code, v.niu_type, v.subtype, v.name,
       COALESCE((SELECT id FROM plugin_sicau_niu_college WHERE "name" = v.college AND "deleted_at" IS NULL), 0),
       v.lat, v.lng, v.status,
       CASE WHEN v.online_now THEN CURRENT_TIMESTAMP - INTERVAL '1 day' ELSE CURRENT_TIMESTAMP + INTERVAL '7 days' END
FROM (VALUES
    ('NIU-001','special','college','信工守护牛','信息工程学院',30.7035,103.8290,'active',true),
    ('NIU-002','special','college','水院奔流牛','水利水电学院',30.7041,103.8302,'active',true),
    ('NIU-003','special','spirit','川农魂','农学院',30.7028,103.8275,'active',true),
    ('NIU-004','special','contribution','奉献牛','动物科技学院',30.7050,103.8311,'active',true),
    ('NIU-005','special','alumni','校友纪念牛','经济学院',30.7019,103.8262,'inactive',false),
    ('NIU-006','common','','望江牛','风景园林学院',30.7063,103.8330,'active',true),
    ('NIU-007','common','','耕读牛','食品学院',30.7008,103.8248,'inactive',false),
    ('NIU-008','common','','勤学牛','林学院',30.7075,103.8345,'active',true),
    ('NIU-009','common','','笃行牛','农学院',30.6998,103.8236,'inactive',false),
    ('NIU-010','common','','至善牛','信息工程学院',30.7088,103.8360,'active',true)
) AS v(code,niu_type,subtype,name,college,lat,lng,status,online_now)
WHERE NOT EXISTS (SELECT 1 FROM plugin_sicau_niu_niu n WHERE n."code" = v.code AND n."deleted_at" IS NULL);

-- 3) 铁牛 (iron) — keyed by code.
INSERT INTO plugin_sicau_niu_iron ("code","name","last_lat","last_lng")
SELECT v.code, v.name, v.lat, v.lng FROM (VALUES
    ('IRON-01','校门铁牛',30.7036,103.8291),
    ('IRON-02','图书馆铁牛',30.7029,103.8276),
    ('IRON-03','体育场铁牛',30.7064,103.8331)
) AS v(code,name,lat,lng)
WHERE NOT EXISTS (SELECT 1 FROM plugin_sicau_niu_iron i WHERE i."code" = v.code AND i."deleted_at" IS NULL);

-- 4) 卡片 (card) — one main card per cattle; niu resolved by code.
INSERT INTO plugin_sicau_niu_card ("niu_id","category","title","content","image_path")
SELECT (SELECT id FROM plugin_sicau_niu_niu WHERE "code" = v.code AND "deleted_at" IS NULL),
       v.category, v.title, v.content, ''
FROM (VALUES
    ('NIU-001','college','信息工程学院','学院以信息技术见长，培养数字川农的中坚力量。'),
    ('NIU-002','college','水利水电学院','治水兴农，水院与都江堰精神一脉相承。'),
    ('NIU-003','spirit','川农大精神','爱国敬业、艰苦奋斗、团结协作、求实创新。'),
    ('NIU-004','person','杰出贡献者','一代代川农人扎根西部、奉献三农。'),
    ('NIU-005','event','建校120周年','百廿川农，弦歌不辍。'),
    ('NIU-006','research','作物科学','作物遗传育种成果服务国家粮食安全。'),
    ('NIU-007','event','耕读传统','耕读结合，知行合一。'),
    ('NIU-008','person','勤学楷模','勤学笃行，止于至善。'),
    ('NIU-009','research','动物营养','畜牧科技助力乡村振兴。'),
    ('NIU-010','event','校史长廊','从相辉堂到温江，川农足迹遍布巴蜀。')
) AS v(code,category,title,content)
WHERE (SELECT id FROM plugin_sicau_niu_niu WHERE "code" = v.code AND "deleted_at" IS NULL) IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM plugin_sicau_niu_card c
    WHERE c."niu_id" = (SELECT id FROM plugin_sicau_niu_niu WHERE "code" = v.code AND "deleted_at" IS NULL)
      AND c."deleted_at" IS NULL);

-- 5) 金句 (quote) — keyed by content; all enabled.
INSERT INTO plugin_sicau_niu_quote ("content","enabled")
SELECT v.content, 1 FROM (VALUES
    ('爱国敬业、艰苦奋斗、团结协作、求实创新'),
    ('追求真理，造福社会，自强不息'),
    ('任重道远，砥砺前行'),
    ('一粒种子可以改变世界'),
    ('扎根西部，服务三农'),
    ('百廿川农，弦歌不辍'),
    ('知行合一，耕读传家'),
    ('为天地立心，为生民立命')
) AS v(content)
WHERE NOT EXISTS (SELECT 1 FROM plugin_sicau_niu_quote q WHERE q."content" = v.content AND q."deleted_at" IS NULL);

-- 6) 荣誉定义 (honor_def) — keyed by code.
INSERT INTO plugin_sicau_niu_honor_def ("honor_type","code","name","unlock_type","threshold","category","sort")
SELECT v.htype, v.code, v.name, v.utype, v.threshold, v.category, v.sort FROM (VALUES
    ('badge','participation_badge','寻牛参与纪念徽章','participation',0,'',1),
    ('badge','feed_bronze','青铜喂养官','feed_count',5,'',2),
    ('badge','feed_silver','白银喂养官','feed_count',15,'',3),
    ('badge','activation_explorer','寻牛探索者','activation_count',3,'',4),
    ('avatar_frame','frame_anniversary','120周年头像框','participation',0,'',5),
    ('badge','category_person','人物收藏徽章','category_complete',0,'person',6),
    ('certificate','cert_participation','川农120周年参与证书','participation',0,'',7),
    ('certificate','cert_full_complete','川农120图鉴收藏证书','full_complete',0,'',8)
) AS v(htype,code,name,utype,threshold,category,sort)
WHERE NOT EXISTS (SELECT 1 FROM plugin_sicau_niu_honor_def h WHERE h."code" = v.code AND h."deleted_at" IS NULL);

-- 7) 玩家 (user) — keyed by openid; college resolved by name. Two players share a
--    device fingerprint (fp-shared) to demonstrate the one-device-many-accounts risk.
INSERT INTO plugin_sicau_niu_user ("openid","phone","nickname","identity_type","college_id","grade","device_fingerprint")
SELECT v.openid, v.phone, v.nickname, v.identity,
       COALESCE((SELECT id FROM plugin_sicau_niu_college WHERE "name" = v.college AND "deleted_at" IS NULL), 0),
       v.grade, v.fp
FROM (VALUES
    ('mock-openid-001','13900000001','信工的牛同学','student','信息工程学院',2023,'fp-001'),
    ('mock-openid-002','13900000002','水院小张','student','水利水电学院',2022,'fp-002'),
    ('mock-openid-003','13900000003','农学阿明','student','农学院',2024,'fp-003'),
    ('mock-openid-004','13900000004','动科李华','student','动物科技学院',2021,'fp-004'),
    ('mock-openid-005','13900000005','园林小美','student','风景园林学院',2023,'fp-005'),
    ('mock-openid-006','13900000006','食院老饕','student','食品学院',2022,'fp-006'),
    ('mock-openid-007','13900000007','九零届校友','alumni','经济学院',2009,'fp-007'),
    ('mock-openid-008','13900000008','王老师','teacher','林学院',0,'fp-008'),
    ('mock-openid-009','13900000009','川农好友老陈','friend','',0,'fp-shared'),
    ('mock-openid-010','13900000010','川农好友小周','friend','',0,'fp-shared')
) AS v(openid,phone,nickname,identity,college,grade,fp)
WHERE NOT EXISTS (SELECT 1 FROM plugin_sicau_niu_user u WHERE u."openid" = v.openid AND u."deleted_at" IS NULL);

-- 8) 草账户 (grass_account) — one per player; keyed by user_id.
INSERT INTO plugin_sicau_niu_grass_account ("user_id","balance")
SELECT u.id, v.balance FROM plugin_sicau_niu_user u
JOIN (VALUES
    ('mock-openid-001',320),('mock-openid-002',180),('mock-openid-003',95),
    ('mock-openid-004',410),('mock-openid-005',60),('mock-openid-006',150),
    ('mock-openid-007',230),('mock-openid-008',80),('mock-openid-009',120),('mock-openid-010',70)
) AS v(openid,balance) ON u."openid" = v.openid AND u."deleted_at" IS NULL
WHERE NOT EXISTS (SELECT 1 FROM plugin_sicau_niu_grass_account a WHERE a."user_id" = u.id AND a."deleted_at" IS NULL);

-- 9) 激活 (activation) — first activators + ordinary; keyed by (user_id, niu_id).
--    activity_date and activated_at are spread over recent days for a meaningful
--    first-activator wall and dashboard.
INSERT INTO plugin_sicau_niu_activation ("user_id","niu_id","activity_date","activated_at","is_first","order_no")
SELECT u.id, n.id, v.adate::date, (v.adate || ' 10:00:00')::timestamp, v.is_first, v.order_no
FROM (VALUES
    ('mock-openid-001','NIU-001', to_char(CURRENT_DATE - 9,'YYYY-MM-DD'), 1, 1),
    ('mock-openid-002','NIU-002', to_char(CURRENT_DATE - 8,'YYYY-MM-DD'), 1, 1),
    ('mock-openid-003','NIU-003', to_char(CURRENT_DATE - 7,'YYYY-MM-DD'), 1, 1),
    ('mock-openid-004','NIU-004', to_char(CURRENT_DATE - 6,'YYYY-MM-DD'), 1, 1),
    ('mock-openid-005','NIU-006', to_char(CURRENT_DATE - 5,'YYYY-MM-DD'), 1, 1),
    ('mock-openid-001','NIU-008', to_char(CURRENT_DATE - 4,'YYYY-MM-DD'), 1, 1),
    ('mock-openid-006','NIU-010', to_char(CURRENT_DATE - 3,'YYYY-MM-DD'), 1, 1),
    ('mock-openid-002','NIU-001', to_char(CURRENT_DATE - 3,'YYYY-MM-DD'), 0, 2),
    ('mock-openid-003','NIU-002', to_char(CURRENT_DATE - 2,'YYYY-MM-DD'), 0, 2),
    ('mock-openid-007','NIU-003', to_char(CURRENT_DATE - 2,'YYYY-MM-DD'), 0, 2),
    ('mock-openid-008','NIU-006', to_char(CURRENT_DATE - 1,'YYYY-MM-DD'), 0, 3),
    ('mock-openid-009','NIU-008', to_char(CURRENT_DATE - 1,'YYYY-MM-DD'), 0, 2)
) AS v(openid,code,adate,is_first,order_no)
JOIN plugin_sicau_niu_user u ON u."openid" = v.openid AND u."deleted_at" IS NULL
JOIN plugin_sicau_niu_niu  n ON n."code"   = v.code   AND n."deleted_at" IS NULL
WHERE NOT EXISTS (
    SELECT 1 FROM plugin_sicau_niu_activation a
    WHERE a."user_id" = u.id AND a."niu_id" = n.id AND a."deleted_at" IS NULL);

-- 10) 喂草 (feeding) — many records over the last 7 days so leaderboards, dashboard
--     and the DAU/retention activity views have data. No unique key: guard on a
--     deterministic (user, niu, base_amount, day-offset) marker via NOT EXISTS on the
--     same shape to stay idempotent.
INSERT INTO plugin_sicau_niu_feeding ("user_id","niu_id","base_amount","coefficient_basis","effect_amount","is_iron_bonus","created_at")
SELECT u.id, n.id, v.base, 100, v.effect, v.iron, (CURRENT_DATE - v.days_ago + TIME '11:00:00')
FROM (VALUES
    ('mock-openid-001','NIU-001',20,30,1,6),
    ('mock-openid-001','NIU-001',20,20,0,5),
    ('mock-openid-001','NIU-003',30,30,0,4),
    ('mock-openid-002','NIU-002',20,20,0,6),
    ('mock-openid-002','NIU-002',15,15,0,3),
    ('mock-openid-003','NIU-003',25,25,0,5),
    ('mock-openid-004','NIU-004',40,60,1,4),
    ('mock-openid-004','NIU-004',30,30,0,2),
    ('mock-openid-005','NIU-006',12,12,0,3),
    ('mock-openid-006','NIU-010',18,18,0,1),
    ('mock-openid-007','NIU-003',22,22,0,2),
    ('mock-openid-008','NIU-006',16,16,0,1),
    ('mock-openid-009','NIU-008',14,14,0,1),
    ('mock-openid-010','NIU-008',13,13,0,1)
) AS v(openid,code,base,effect,iron,days_ago)
JOIN plugin_sicau_niu_user u ON u."openid" = v.openid AND u."deleted_at" IS NULL
JOIN plugin_sicau_niu_niu  n ON n."code"   = v.code   AND n."deleted_at" IS NULL
WHERE NOT EXISTS (
    SELECT 1 FROM plugin_sicau_niu_feeding f
    WHERE f."user_id" = u.id AND f."niu_id" = n.id
      AND f."base_amount" = v.base AND f."effect_amount" = v.effect
      AND f."created_at"::date = (CURRENT_DATE - v.days_ago) AND f."deleted_at" IS NULL);

-- 11) 签到 (checkin) — keyed by (user_id, checkin_date); a few recent days each.
INSERT INTO plugin_sicau_niu_checkin ("user_id","checkin_date","amount")
SELECT u.id, (CURRENT_DATE - v.days_ago)::text, v.amount
FROM (VALUES
    ('mock-openid-001',0,35),('mock-openid-001',1,28),('mock-openid-001',2,42),
    ('mock-openid-002',0,30),('mock-openid-002',1,25),
    ('mock-openid-003',0,48),('mock-openid-004',0,22),('mock-openid-005',1,33),
    ('mock-openid-006',0,40),('mock-openid-007',0,20)
) AS v(openid,days_ago,amount)
JOIN plugin_sicau_niu_user u ON u."openid" = v.openid AND u."deleted_at" IS NULL
WHERE NOT EXISTS (
    SELECT 1 FROM plugin_sicau_niu_checkin c
    WHERE c."user_id" = u.id AND c."checkin_date" = (CURRENT_DATE - v.days_ago)::text AND c."deleted_at" IS NULL);

-- 12) 偷草 (steal) — actor/target by openid; guard on (actor, target, steal_date).
INSERT INTO plugin_sicau_niu_steal ("actor_user_id","target_user_id","amount","steal_date")
SELECT a.id, t.id, v.amount, (CURRENT_DATE - v.days_ago)::text
FROM (VALUES
    ('mock-openid-001','mock-openid-002',8,1),
    ('mock-openid-002','mock-openid-003',5,1),
    ('mock-openid-003','mock-openid-004',6,2),
    ('mock-openid-009','mock-openid-010',4,0),
    ('mock-openid-004','mock-openid-001',7,0)
) AS v(actor,target,amount,days_ago)
JOIN plugin_sicau_niu_user a ON a."openid" = v.actor  AND a."deleted_at" IS NULL
JOIN plugin_sicau_niu_user t ON t."openid" = v.target AND t."deleted_at" IS NULL
WHERE NOT EXISTS (
    SELECT 1 FROM plugin_sicau_niu_steal s
    WHERE s."actor_user_id" = a.id AND s."target_user_id" = t.id
      AND s."steal_date" = (CURRENT_DATE - v.days_ago)::text AND s."deleted_at" IS NULL);

-- 13) 送草 (gift) — from/to by openid; guard on (from, to, gift_date).
INSERT INTO plugin_sicau_niu_gift ("from_user_id","to_user_id","amount","gift_date")
SELECT f.id, t.id, v.amount, (CURRENT_DATE - v.days_ago)::text
FROM (VALUES
    ('mock-openid-001','mock-openid-005',12,1),
    ('mock-openid-004','mock-openid-006',24,1),
    ('mock-openid-002','mock-openid-003',12,0)
) AS v(fromid,toid,amount,days_ago)
JOIN plugin_sicau_niu_user f ON f."openid" = v.fromid AND f."deleted_at" IS NULL
JOIN plugin_sicau_niu_user t ON t."openid" = v.toid   AND t."deleted_at" IS NULL
WHERE NOT EXISTS (
    SELECT 1 FROM plugin_sicau_niu_gift g
    WHERE g."from_user_id" = f.id AND g."to_user_id" = t.id
      AND g."gift_date" = (CURRENT_DATE - v.days_ago)::text AND g."deleted_at" IS NULL);

-- 14) 玩家荣誉授予 (user_honor) — keyed by (user_id, honor_id); user/honor by business key.
INSERT INTO plugin_sicau_niu_user_honor ("user_id","honor_id","unlocked_at")
SELECT u.id, h.id, NOW()
FROM (VALUES
    ('mock-openid-001','participation_badge'),
    ('mock-openid-001','feed_bronze'),
    ('mock-openid-001','cert_participation'),
    ('mock-openid-002','participation_badge'),
    ('mock-openid-002','cert_participation'),
    ('mock-openid-003','participation_badge'),
    ('mock-openid-004','feed_bronze'),
    ('mock-openid-004','cert_participation')
) AS v(openid,code)
JOIN plugin_sicau_niu_user u      ON u."openid" = v.openid AND u."deleted_at" IS NULL
JOIN plugin_sicau_niu_honor_def h ON h."code"   = v.code   AND h."deleted_at" IS NULL
WHERE NOT EXISTS (
    SELECT 1 FROM plugin_sicau_niu_user_honor uh
    WHERE uh."user_id" = u.id AND uh."honor_id" = h.id AND uh."deleted_at" IS NULL);

-- 15) 结算归档 (settlement) — one demo snapshot; guard on title.
INSERT INTO plugin_sicau_niu_settlement ("title","snapshot","operator_id","archived_at")
SELECT '寻牛活动结算公示（演示）',
       '{"playerCount":10,"activatedNiuCount":6,"firstActivatorCount":7,"certificateGrantedCount":4}',
       0, NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM plugin_sicau_niu_settlement s
    WHERE s."title" = '寻牛活动结算公示（演示）' AND s."deleted_at" IS NULL);
