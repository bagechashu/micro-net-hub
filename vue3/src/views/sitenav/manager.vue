<template>
  <div>
    <el-card class="container-card" shadow="always">
      <el-form
        ref="navGroupFormRef"
        :inline="true"
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
          <el-button :loading="loading" type="primary" @click="addGroup()">
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
          <el-form :inline="true" class="demo-form-inline">
            <el-form-item>
              <el-button :loading="loading" :icon="Plus" type="primary" @click="addSite">
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
            stripe
            style="width: 100%"
            @selection-change="handleSelectionChange"
          >
            <el-table-column type="selection" width="48" align="center" />
            <el-table-column show-overflow-tooltip sortable prop="name" :label="$t('sitenav.6112v35e63c0')" min-width="140" />
            <el-table-column prop="icon" :label="$t('sitenav.6112v35e66c0')" width="90">
              <template #default="scope">
                <el-tooltip :content="scope.row.icon || '—'" placement="top">
                  <SiteIcon :value="scope.row.icon" :name="scope.row.name" :link="scope.row.link" :size="28" />
                </el-tooltip>
              </template>
            </el-table-column>
            <el-table-column show-overflow-tooltip sortable prop="desc" :label="$t('sitenav.6112v35e68w0')" min-width="160" />
            <el-table-column show-overflow-tooltip sortable prop="link" :label="$t('sitenav.6112v35e6bw0')" min-width="220" />
            <el-table-column show-overflow-tooltip sortable prop="doc" :label="$t('sitenav.6112v35e6eo0')" min-width="160" />
            <el-table-column fixed="right" :label="$t('sitenav.6112v35e6ho0')" align="left" width="160">
              <template #default="scope">
                <div class="row-actions">
                  <el-button link type="primary" :icon="Edit" @click="updateSite(scope.row)">{{ $t('common.edit') }}</el-button>
                  <el-popconfirm :title="$t('sitenav.6112v35e6pk0')" @confirm="deleteSite(scope.row.ID)">
                    <template #reference>
                      <el-button link type="danger" :icon="Delete">{{ $t('common.delete') }}</el-button>
                    </template>
                  </el-popconfirm>
                </div>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
      <el-empty v-else :description="$t('common.nodata')" />

      <el-dialog v-model="navSiteFormVisible" :title="navSiteFormTitle" @closed="resetSiteFormFields">
        <el-form
          ref="navSiteFormRef"
          :model="navSiteForm"
          :rules="navSiteFormRules"
          label-position="top"
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
          <el-form-item :label="$t('sitenav.6112v35e6bw0')" prop="link">
            <el-input
              v-model="navSiteForm.link"
              :placeholder="$t('sitenav.6112v35e6bw0')"
              @blur="navSiteForm.link = navSiteForm.link.trim()"
            />
          </el-form-item>
          <el-form-item :label="$t('sitenav.6112v35e66c0')" prop="icon">
            <IconPicker
              v-model="navSiteForm.icon"
              :name="navSiteForm.name"
              :link="navSiteForm.link"
              @update:model-value="navSiteFormRef?.validateField('icon')"
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
            <el-button @click="navSiteFormCancel()">{{ $t('sitenav.6112v35e74g0') }}</el-button>
            <el-button :loading="loading" type="primary" @click="navSiteFormSubmit()">
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
import { Plus, Delete, Edit } from '@/icons/tabler';
import SiteIcon from '@/components/SiteIcon/index.vue';
import IconPicker from '@/components/SiteIcon/IconPicker.vue';
import { isValidSiteIcon } from '@/components/SiteIcon/icon';
import {
  getNav,
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
 * Absolute http(s) URL check. The WHATWG URL parser is used so that tricks like
 * " javascript:" or "java\tscript:" are normalised the same way a browser would.
 */
function isSafeHttpUrl(value: string): boolean {
  const v = (value || '').trim();
  if (!v) return false;
  try {
    const u = new URL(v);
    return u.protocol === 'http:' || u.protocol === 'https:';
  } catch {
    return false;
  }
}

function urlRule(): FormItemRule {
  return {
    trigger: 'blur',
    validator: (_rule, value, callback) => {
      if (!value || !String(value).trim()) return callback(); // emptiness handled by `required`
      if (!isSafeHttpUrl(String(value))) return callback(new Error(invalidUrlMessage()));
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
  // 图标可选：si:/ti:/favicon/图片地址，留空显示首字头像
  icon: [
    { max: 100, message: t('valid.length', [0, 100]), trigger: 'change' },
    {
      trigger: 'change',
      validator: (_rule, value, callback) => (isValidSiteIcon(value) ? callback() : callback(new Error(t('sitenav.iconPicker.invalid'))))
    }
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
    isSafeHttpUrl(form.link) && (!form.doc || isSafeHttpUrl(form.doc)) && isValidSiteIcon(form.icon);
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
