# Test Data of SiteNav

`icon_url` 的写法（前端 `frontend/src/components/SiteIcon/icon.ts`）：

- `si:<slug>`：品牌 Logo，如 `si:jenkins`、`si:gitlab`（simple-icons 精选）
- `ti:<name>[:<color>]`：Tabler 通用图标，如 `ti:server`、`ti:database:green`
- `favicon`：自动使用站点自己的 favicon
- `http(s)://…`：图片地址
- 留空：显示站点名称首字

```sql
-- Insert data for 'company' NavGroup
INSERT INTO nav_groups (title, name) VALUES ('公司环境地址', 'company');
SET @company_id := LAST_INSERT_ID();

-- Insert data for 'company' NavSites
INSERT INTO nav_sites (nav_group_id, icon_url, name, description, link, doc_url) VALUES
(@company_id, 'si:jenkins', 'jenkins', '测试环境Jenkins', 'http://127.0.0.1:8080/', '/data/docs/jenkins.md'),
(@company_id, 'si:jenkins', 'jenkins1', '测试环境Jenkins', 'http://127.0.0.1:8080/', '/data/docs/jenkins.md'),
(@company_id, 'si:jenkins', 'jenkins2', '测试环境Jenkins', 'http://127.0.0.1:8080/', '/data/docs/jenkins.md'),
(@company_id, 'si:jenkins', 'jenkins3', '测试环境Jenkins', 'http://127.0.0.1:8080/', '/data/docs/jenkins.md'),
(@company_id, 'si:jenkins', 'jenkins4', '测试环境Jenkins', 'http://127.0.0.1:8080/', '/data/docs/jenkins.md'),
(@company_id, 'si:jenkins', 'jenkins5', '测试环境Jenkins', 'http://127.0.0.1:8080/', '/data/docs/jenkins.md'),
(@company_id, 'si:jenkins', 'jenkins6', '测试环境Jenkins', 'http://127.0.0.1:8080/', '/data/docs/jenkins.md'),
(@company_id, 'si:jenkins', 'jenkins7', '测试环境Jenkins', 'http://127.0.0.1:8080/', '/data/docs/jenkins.md'),
(@company_id, 'si:confluence', 'Confluence', 'Confluence, 技术文档', 'http://127.0.0.1:8090/', 'https://www.atlassian.com/software/confluence');
-- Add more rows for other NavSites in the 'company' group

-- Insert data for 'test1' NavGroup
INSERT INTO nav_groups (title, name) VALUES ('Test1环境地址', 'test1');
SET @test1_id := LAST_INSERT_ID();

-- Insert data for 'test1' NavSites
INSERT INTO nav_sites (nav_group_id, icon_url, name, description, link) VALUES
(@test1_id, 'si:confluence', 'test1', 'Test1环境详细信息', 'http://127.0.0.1:8090/pages/viewpage.action?pageId=6127905');
-- Add more rows for other NavSites in the 'test1' group

-- Repeat similar INSERT statements for 'test2', 'test3', 'pre', and 'prod' NavSideGroups and NavGroups

INSERT INTO nav_groups (title, name) VALUES ('Test2环境地址', 'test2');
INSERT INTO nav_groups (title, name) VALUES ('Test3环境地址', 'test3');
-- Finally, commit the transaction
COMMIT;


```
