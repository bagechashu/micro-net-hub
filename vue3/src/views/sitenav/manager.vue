<template>
  <div>
    <el-card class="container-card" shadow="always">
      <el-form
        ref="navGroupFormRef"
        :inline="true"
        size="small"
        :model="navGroupForm"
        :rules="navGroupFormRules"
      >
        <el-form-item :label="$t('sitenav.6112v35e5140')" prop="name">
          <el-input v-model.trim="navGroupForm.name" :placeholder="$t('sitenav.6112v35e5ik0')" />
        </el-form-item>
        <el-form-item :label="$t('sitenav.6112v35e5o40')" prop="title">
          <el-input
            v-model="navGroupForm.title"
            :placeholder="$t('sitenav.6112v35e5r40')"
            @blur="navGroupForm.title = navGroupForm.title.trim()"
          />
        </el-form-item>
        <el-form-item>
          <el-button size="small" :loading="loading" type="primary" @click="addGroup()">
            {{ $t('sitenav.6112v35e5v40') }}
          </el-button>
        </el-form-item>
      </el-form>

      <el-tabs
        v-if="navData.length > 0"
        v-model="navGroupActiveTab"
        type="border-card"
        closable
        @tab-remove="deleteGroup"
      >
        <el-tab-pane v-for="item in navData" :key="item.name" :label="item.title" :name="item.name">
          <el-form size="small" :inline="true" class="demo-form-inline">
            <el-form-item>
              <el-button :loading="loading" :icon="Plus" type="warning" @click="addSite">
                {{ $t('sitenav.6112v35e5xw0') }}
              </el-button>
            </el-form-item>
            <el-form-item>
              <el-button
                :disabled="multipleSelection.length === 0"
                :loading="loading"
                :icon="Delete"
                type="danger"
                @click="batchDeleteSites"
              >
                {{ $t('sitenav.6112v35e60s0') }}
              </el-button>
            </el-form-item>
          </el-form>

          <el-table
            v-loading="loading"
            :data="item.sites"
            border
            stripe
            size="small"
            style="width: 100%"
            @selection-change="handleSelectionChange"
          >
            <el-table-column type="selection" width="55" align="center" />
            <el-table-column show-overflow-tooltip sortable prop="name" :label="$t('sitenav.6112v35e63c0')" />
            <el-table-column show-overflow-tooltip sortable prop="icon" :label="$t('sitenav.6112v35e66c0')" />
            <el-table-column show-overflow-tooltip sortable prop="desc" :label="$t('sitenav.6112v35e68w0')" />
            <el-table-column show-overflow-tooltip sortable prop="link" :label="$t('sitenav.6112v35e6bw0')" />
            <el-table-column show-overflow-tooltip sortable prop="doc" :label="$t('sitenav.6112v35e6eo0')" />
            <el-table-column fixed="right" :label="$t('sitenav.6112v35e6ho0')" align="center" width="120">
              <template #default="scope">
                <el-tooltip :content="$t('sitenav.6112v35e6k80')" effect="dark" placement="top">
                  <el-button size="small" :icon="Edit" circle type="primary" @click="updateSite(scope.row)" />
                </el-tooltip>
                <el-tooltip class="delete-popover" :content="$t('sitenav.6112v35e6mw0')" effect="dark" placement="top">
                  <template #default>
                    <span>
                      <el-popconfirm :title="$t('sitenav.6112v35e6pk0')" @confirm="deleteSite(scope.row.ID)">
                        <template #reference>
                          <el-button size="small" :icon="Delete" circle type="danger" />
                        </template>
                      </el-popconfirm>
                    </span>
                  </template>
                </el-tooltip>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
      <el-empty v-else :description="$t('common.nodata')" />

      <el-dialog v-model="navSiteFormVisible" :title="navSiteFormTitle" @closed="resetSiteFormFields">
        <el-form
          ref="navSiteFormRef"
          size="small"
          :model="navSiteForm"
          :rules="navSiteFormRules"
          label-width="auto"
        >
          <el-form-item :label="$t('sitenav.6112v35e6s80')" prop="groupid">
            <el-select v-model="navSiteForm.groupid" :placeholder="$t('sitenav.6112v35e6v40')">
              <el-option v-for="item in groupOptions" :key="item.id" :label="item.name" :value="item.id" />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('sitenav.6112v35e63c0')" prop="name">
            <el-input
              v-model="navSiteForm.name"
              :placeholder="$t('sitenav.6112v35e63c0')"
              @blur="navSiteForm.name = navSiteForm.name.trim()"
            />
          </el-form-item>
          <el-form-item :label="$t('sitenav.6112v35e66c0')" prop="icon">
            <el-select
              v-model="navSiteForm.icon"
              filterable
              allow-create
              default-first-option
              :placeholder="$t('sitenav.6112v35e6y00')"
              @blur="navSiteForm.icon = (navSiteForm.icon || '').trim()"
            >
              <el-option v-for="item in iconOptions" :key="item.url" :label="item.url" :value="item.url" />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('sitenav.6112v35e6bw0')" prop="link">
            <el-input
              v-model="navSiteForm.link"
              :placeholder="$t('sitenav.6112v35e6bw0')"
              @blur="navSiteForm.link = navSiteForm.link.trim()"
            />
          </el-form-item>
          <el-form-item :label="$t('sitenav.6112v35e6eo0')" prop="doc">
            <el-input
              v-model="navSiteForm.doc"
              :placeholder="$t('sitenav.6112v35e70o0')"
              @blur="navSiteForm.doc = navSiteForm.doc.trim()"
            />
          </el-form-item>
          <el-form-item :label="$t('sitenav.6112v35e68w0')" prop="desc">
            <el-input
              v-model="navSiteForm.desc"
              type="textarea"
              :placeholder="$t('sitenav.6112v35e68w0')"
              show-word-limit
              maxlength="100"
              @blur="navSiteForm.desc = navSiteForm.desc.trim()"
            />
          </el-form-item>
        </el-form>
        <template #footer>
          <div class="dialog-footer">
            <el-button size="small" @click="navSiteFormCancel()">{{ $t('sitenav.6112v35e74g0') }}</el-button>
            <el-button size="small" :loading="loading" type="primary" @click="navSiteFormSubmit()">
              {{ $t('sitenav.6112v35e76o0') }}
            </el-button>
          </div>
        </template>
      </el-dialog>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { ElMessage, ElMessageBox } from 'element-plus';
import type { FormInstance, FormItemRule } from 'element-plus';
import { Plus, Delete, Edit } from '@element-plus/icons-vue';
import {
  getNav,
  getIcons,
  createNavGroup,
  batchDeleteNavGroupByIds,
  createNavSite,
  updateNavSite,
  batchDeleteNavSiteByIds
} from '@/api/sitenav';

defineOptions({ name: 'SiteManager' });

const { t, te, locale } = useI18n();

const loading = ref(false);
const navData = ref<any[]>([]);

const navGroupFormRef = ref<FormInstance>();
const navGroupForm = reactive({ name: '', title: '' });
const navGroupFormRules = {
  name: [
    { required: true, message: t('sitenav.6112v35e7980'), trigger: 'blur' },
    { min: 4, max: 20, message: t('valid.length', [4, 20]), trigger: 'blur' }
  ],
  title: [
    { required: true, message: t('sitenav.6112v35e7bc0'), trigger: 'blur' },
    { min: 4, max: 20, message: t('valid.length', [4, 20]), trigger: 'blur' }
  ]
};
const navGroupActiveTab = ref<string | number>('');

// ---- URL scheme validation (only http/https are allowed, blocks javascript:/data: etc.) ----
function invalidUrlMessage() {
  if (te('valid.invalidUrlScheme')) return t('valid.invalidUrlScheme');
  return String(locale.value).startsWith('zh')
    ? '仅支持 http:// 或 https:// 开头的地址'
    : 'Only http:// or https:// URLs are allowed';
}

/**
 * Absolute http(s) URL check. `allowRootRelative` additionally accepts same-origin
 * paths such as "/ui/assets/logo/xx.png" (used by the built-in icon list).
 * The WHATWG URL parser is used so that tricks like " javascript:" or "java\tscript:"
 * are normalised the same way a browser would before the protocol check.
 */
function isSafeHttpUrl(value: string, allowRootRelative = false): boolean {
  const v = (value || '').trim();
  if (!v) return false;
  try {
    if (allowRootRelative && v.startsWith('/') && !v.startsWith('//') && !v.startsWith('/\\')) {
      const u = new URL(v, window.location.origin);
      return u.origin === window.location.origin;
    }
    const u = new URL(v);
    return u.protocol === 'http:' || u.protocol === 'https:';
  } catch {
    return false;
  }
}

function urlRule(allowRootRelative = false): FormItemRule {
  return {
    trigger: 'blur',
    validator: (_rule, value, callback) => {
      if (!value || !String(value).trim()) return callback(); // emptiness handled by `required`
      if (!isSafeHttpUrl(String(value), allowRootRelative)) return callback(new Error(invalidUrlMessage()));
      callback();
    }
  };
}

const navSiteFormRef = ref<FormInstance>();
const navSiteFormTitle = ref('');
const navSiteFormType = ref('');
const navSiteFormVisible = ref(false);
const emptySiteForm = () => ({
  ID: 0,
  name: '',
  icon: '',
  link: '',
  doc: '',
  desc: '',
  groupid: 1 as number | undefined
});
const navSiteForm = ref(emptySiteForm());
const navSiteFormRules = {
  name: [
    { required: true, message: t('sitenav.6112v35e7dc0'), trigger: 'blur' },
    { min: 1, max: 50, message: t('valid.length', [1, 50]), trigger: 'blur' }
  ],
  icon: [
    { required: true, message: t('sitenav.6112v35e7fw0'), trigger: 'blur' },
    { min: 1, max: 100, message: t('valid.length', [1, 100]), trigger: 'blur' },
    urlRule(true)
  ],
  link: [
    { required: true, message: t('sitenav.6112v35e7hw0'), trigger: 'change' },
    { min: 0, max: 100, message: t('valid.length', [1, 100]), trigger: 'blur' },
    urlRule()
  ],
  doc: [
    { required: false, message: t('sitenav.6112v35e7k00'), trigger: 'blur' },
    { min: 0, max: 100, message: t('valid.length', [1, 100]), trigger: 'blur' },
    urlRule()
  ],
  desc: [
    { required: true, message: t('sitenav.6112v35e7m40'), trigger: 'blur' },
    { min: 0, max: 200, message: t('valid.length', [1, 200]), trigger: 'blur' }
  ],
  groupid: [{ required: true, message: t('sitenav.6112v35e7o40'), trigger: 'blur' }]
};
const iconOptions = ref<{ url: string }[]>([]);
const groupOptions = ref<{ id: number; name: string }[]>([]);
const multipleSelection = ref<any[]>([]);

async function getData() {
  loading.value = true;
  try {
    const { data } = await getNav();
    navData.value = data || [];
    if ((navGroupActiveTab.value === '0' || navGroupActiveTab.value === '') && navData.value.length > 0) {
      navGroupActiveTab.value = navData.value[0].name;
    }
  } catch {
    /* already shown */
  } finally {
    loading.value = false;
  }
}

async function getIconOptions() {
  try {
    const res: any = await getIcons();
    iconOptions.value = res.logos || [];
  } catch {
    /* already shown */
  }
}

function getGroupOptions() {
  groupOptions.value = navData.value.map(item => ({ id: item.ID, name: item.title }));
}

function successMessage() {
  ElMessage({ showClose: true, message: t('sitenav.6112v35e8140'), type: 'success' });
}

function validateFailedMessage() {
  ElMessage({ showClose: true, message: t('sitenav.6112v35e7qg0'), type: 'warning' });
}

function cancelledMessage() {
  ElMessage({ showClose: true, type: 'info', message: t('sitenav.6112v35e7uw0') });
}

async function addGroup() {
  const valid = await navGroupFormRef.value?.validate().catch(() => false);
  if (!valid) {
    validateFailedMessage();
    return;
  }
  loading.value = true;
  try {
    await createNavGroup({ ...navGroupForm });
    successMessage();
  } catch {
    /* already shown */
  } finally {
    loading.value = false;
  }
  getData();
}

function getGroupIDFromTabname(tabname: string | number) {
  return navData.value.find(item => item.name === tabname)?.ID;
}

async function deleteGroup(tabname: string | number) {
  const navGroupId = getGroupIDFromTabname(tabname);
  try {
    await ElMessageBox.confirm(t('sitenav.deleteSiteTips'), t('sitenav.6112v35e7sg0'), {
      confirmButtonText: t('sitenav.6112v35e76o0'),
      cancelButtonText: t('sitenav.6112v35e74g0'),
      type: 'warning'
    });
  } catch {
    cancelledMessage();
    return;
  }
  loading.value = true;
  try {
    await batchDeleteNavGroupByIds({ ids: [navGroupId] });
    successMessage();
    navGroupActiveTab.value = '0';
  } catch {
    /* already shown */
  } finally {
    loading.value = false;
  }
  getData();
}

function addSite() {
  getGroupOptions();
  navSiteForm.value.groupid = getGroupIDFromTabname(navGroupActiveTab.value);
  navSiteFormTitle.value = t('sitenav.6112v35e7ww0');
  navSiteFormType.value = 'add';
  navSiteFormVisible.value = true;
}

function updateSite(row: any) {
  getGroupOptions();
  navSiteForm.value = {
    ID: row.ID,
    name: row.name,
    icon: row.icon,
    desc: row.desc,
    link: row.link,
    doc: row.doc,
    groupid: row.groupid
  };
  navSiteFormTitle.value = t('sitenav.6112v35e7z00');
  navSiteFormType.value = 'update';
  navSiteFormVisible.value = true;
}

async function deleteSite(id: number) {
  loading.value = true;
  try {
    await batchDeleteNavSiteByIds({ ids: [id] });
    successMessage();
  } catch {
    /* already shown */
  } finally {
    loading.value = false;
  }
  getData();
}

async function batchDeleteSites() {
  try {
    await ElMessageBox.confirm(t('tips.deleteWarning'), t('sitenav.6112v35e7sg0'), {
      confirmButtonText: t('sitenav.6112v35e76o0'),
      cancelButtonText: t('sitenav.6112v35e74g0'),
      type: 'warning'
    });
  } catch {
    cancelledMessage();
    return;
  }
  loading.value = true;
  try {
    await batchDeleteNavSiteByIds({ ids: multipleSelection.value.map(x => x.ID) });
    successMessage();
  } catch {
    /* already shown */
  } finally {
    loading.value = false;
  }
  getData();
}

async function navSiteFormSubmit() {
  const form = navSiteForm.value;
  form.name = form.name.trim();
  form.icon = (form.icon || '').trim();
  form.link = form.link.trim();
  form.doc = form.doc.trim();
  form.desc = form.desc.trim();

  const valid = await navSiteFormRef.value?.validate().catch(() => false);
  // Defence in depth: never submit a non-http(s) link/doc/icon even if the rules were bypassed.
  const urlsOk =
    isSafeHttpUrl(form.link) && (!form.doc || isSafeHttpUrl(form.doc)) && isSafeHttpUrl(form.icon, true);
  if (!valid || !urlsOk) {
    validateFailedMessage();
    return;
  }
  loading.value = true;
  try {
    if (navSiteFormType.value === 'add') {
      await createNavSite({ ...form });
    } else if (navSiteFormType.value === 'update') {
      await updateNavSite({ ...form });
    }
    successMessage();
  } catch {
    loading.value = false;
    return;
  }
  loading.value = false;
  navSiteFormReset();
  getData();
}

function navSiteFormCancel() {
  navSiteFormReset();
}

function resetSiteFormFields() {
  navSiteFormRef.value?.resetFields();
  navSiteForm.value = emptySiteForm();
}

function navSiteFormReset() {
  navSiteFormVisible.value = false;
  resetSiteFormFields();
}

function handleSelectionChange(val: any[]) {
  multipleSelection.value = val;
}

onMounted(() => {
  getData();
  getIconOptions();
});
</script>

<style scoped>
.container-card {
  margin: 10px;
  margin-bottom: 100px;
}

.delete-popover {
  margin-left: 10px;
}
</style>
