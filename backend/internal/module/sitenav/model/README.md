# Test Data of SiteNav

```sql
-- Insert data for 'company' NavGroup
INSERT INTO nav_groups (title, name) VALUES ('公司环境地址', 'company');
SET @company_id := LAST_INSERT_ID();

-- Insert data for 'company' NavSites
INSERT INTO nav_sites (nav_group_id, icon_url, name, description, link, doc_url) VALUES
(@company_id, 'http://127.0.0.1:9000/ui/assets/logo/jenkins.png', 'jenkins', '测试环境Jenkins', 'http://127.0.0.1:8080/', '/data/docs/jenkins.md'),
(@company_id, 'http://127.0.0.1:9000/ui/assets/logo/jenkins.png', 'jenkins1', '测试环境Jenkins', 'http://127.0.0.1:8080/', '/data/docs/jenkins.md'),
(@company_id, 'http://127.0.0.1:9000/ui/assets/logo/jenkins.png', 'jenkins2', '测试环境Jenkins', 'http://127.0.0.1:8080/', '/data/docs/jenkins.md'),
(@company_id, 'http://127.0.0.1:9000/ui/assets/logo/jenkins.png', 'jenkins3', '测试环境Jenkins', 'http://127.0.0.1:8080/', '/data/docs/jenkins.md'),
(@company_id, 'http://127.0.0.1:9000/ui/assets/logo/jenkins.png', 'jenkins4', '测试环境Jenkins', 'http://127.0.0.1:8080/', '/data/docs/jenkins.md'),
(@company_id, 'http://127.0.0.1:9000/ui/assets/logo/jenkins.png', 'jenkins5', '测试环境Jenkins', 'http://127.0.0.1:8080/', '/data/docs/jenkins.md'),
(@company_id, 'http://127.0.0.1:9000/ui/assets/logo/jenkins.png', 'jenkins6', '测试环境Jenkins', 'http://127.0.0.1:8080/', '/data/docs/jenkins.md'),
(@company_id, 'http://127.0.0.1:9000/ui/assets/logo/jenkins.png', 'jenkins7', '测试环境Jenkins', 'http://127.0.0.1:8080/', '/data/docs/jenkins.md'),
(@company_id, '/ui/assets/logo/confluence.png', 'Confluence', 'Confluence, 技术文档', 'http://127.0.0.1:8090/', 'https://www.atlassian.com/software/confluence');
-- Add more rows for other NavSites in the 'company' group

-- Insert data for 'test1' NavGroup
INSERT INTO nav_groups (title, name) VALUES ('Test1环境地址', 'test1');
SET @test1_id := LAST_INSERT_ID();

-- Insert data for 'test1' NavSites
INSERT INTO nav_sites (nav_group_id, icon_url, name, description, link) VALUES
(@test1_id, '/ui/assets/logo/confluence.png', 'test1', 'Test1环境详细信息', 'http://127.0.0.1:8090/pages/viewpage.action?pageId=6127905');
-- Add more rows for other NavSites in the 'test1' group

-- Repeat similar INSERT statements for 'test2', 'test3', 'pre', and 'prod' NavSideGroups and NavGroups

INSERT INTO nav_groups (title, name) VALUES ('Test2环境地址', 'test2');
INSERT INTO nav_groups (title, name) VALUES ('Test3环境地址', 'test3');
-- Finally, commit the transaction
COMMIT;


```
