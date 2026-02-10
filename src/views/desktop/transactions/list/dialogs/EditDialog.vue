<template>
    <v-dialog width="1000" :persistent="isTransactionModified" v-model="showState">
        <v-card class="pa-sm-1 pa-md-2">
            <template #title>
                <div class="d-flex align-center justify-center">
                    <div class="d-flex align-center">
                        <h4 class="text-h4">{{ tt(title) }}</h4>
                        <v-chip class="ms-2" color="warning" size="small" v-if="transaction.planned">{{ tt('Planned') }}</v-chip>
                        <v-progress-circular indeterminate size="22" class="ms-2" v-if="loading"></v-progress-circular>
                    </div>
                    <v-spacer/>
                    <v-btn density="comfortable" color="default" variant="text" class="ms-2" :icon="true"
                           :disabled="loading || submitting" v-if="mode !== TransactionEditPageMode.View && (activeTab === 'basicInfo' || (activeTab === 'map' && isSupportGetGeoLocationByClick()))">
                        <v-icon :icon="mdiDotsVertical" />
                        <v-menu activator="parent">
                            <v-list v-if="activeTab === 'basicInfo'">
                                <v-list-item :prepend-icon="mdiSwapHorizontal"
                                             :title="tt('Swap Account')"
                                             v-if="transaction.type === TransactionType.Transfer"
                                             @click="swapTransactionData(true, false)"></v-list-item>
                                <v-list-item :prepend-icon="mdiSwapHorizontal"
                                             :title="tt('Swap Amount')"
                                             v-if="transaction.type === TransactionType.Transfer"
                                             @click="swapTransactionData(false, true)"></v-list-item>
                                <v-list-item :prepend-icon="mdiSwapHorizontal"
                                             :title="tt('Swap Account and Amount')"
                                             v-if="transaction.type === TransactionType.Transfer"
                                             @click="swapTransactionData(true, true)"></v-list-item>
                                <v-divider v-if="transaction.type === TransactionType.Transfer" />
                                <v-list-item :prepend-icon="mdiEyeOutline"
                                             :title="tt('Show Amount')"
                                             v-if="transaction.hideAmount" @click="transaction.hideAmount = false"></v-list-item>
                                <v-list-item :prepend-icon="mdiEyeOffOutline"
                                             :title="tt('Hide Amount')"
                                             v-if="!transaction.hideAmount" @click="transaction.hideAmount = true"></v-list-item>
                            </v-list>
                            <v-list v-if="activeTab === 'map'">
                                <v-list-item key="setGeoLocationByClickMap" value="setGeoLocationByClickMap"
                                             :prepend-icon="mdiMapMarkerOutline"
                                             :disabled="!transaction.geoLocation" v-if="isSupportGetGeoLocationByClick()">
                                    <v-list-item-title class="cursor-pointer" @click="setGeoLocationByClickMap = !setGeoLocationByClickMap; geoMenuState = false">
                                        <div class="d-flex align-center">
                                            <span>{{ tt('Click on Map to Set Geographic Location') }}</span>
                                            <v-spacer/>
                                            <v-icon :icon="mdiCheck" v-if="setGeoLocationByClickMap" />
                                        </div>
                                    </v-list-item-title>
                                </v-list-item>
                            </v-list>
                        </v-menu>
                    </v-btn>
                </div>
            </template>
            <v-card-text class="d-flex flex-column flex-md-row flex-grow-1 overflow-y-auto">
                <div class="mb-4">
                    <v-tabs class="v-tabs-pill" direction="vertical" :class="{ 'readonly': type === TransactionEditPageType.Transaction && mode !== TransactionEditPageMode.Add }"
                            :disabled="loading || submitting" v-model="transaction.type">
                        <v-tab :value="TransactionType.Expense" :disabled="type === TransactionEditPageType.Transaction && mode !== TransactionEditPageMode.Add && transaction.type !== TransactionType.Expense" v-if="transaction.type !== TransactionType.ModifyBalance">
                            <span>{{ tt('Expense') }}</span>
                        </v-tab>
                        <v-tab :value="TransactionType.Income" :disabled="type === TransactionEditPageType.Transaction && mode !== TransactionEditPageMode.Add && transaction.type !== TransactionType.Income" v-if="transaction.type !== TransactionType.ModifyBalance">
                            <span>{{ tt('Income') }}</span>
                        </v-tab>
                        <v-tab :value="TransactionType.Transfer" :disabled="type === TransactionEditPageType.Transaction && mode !== TransactionEditPageMode.Add && transaction.type !== TransactionType.Transfer" v-if="transaction.type !== TransactionType.ModifyBalance">
                            <span>{{ tt('Transfer') }}</span>
                        </v-tab>
                        <v-tab :value="TransactionType.ModifyBalance" v-if="type === TransactionEditPageType.Transaction && transaction.type === TransactionType.ModifyBalance">
                            <span>{{ tt('Modify Balance') }}</span>
                        </v-tab>
                    </v-tabs>
                    <v-divider class="my-2"/>
                    <v-tabs direction="vertical" :disabled="loading || submitting" v-model="activeTab">
                        <v-tab value="basicInfo">
                            <span>{{ tt('Basic Information') }}</span>
                        </v-tab>
                        <!-- Hidden: Location on Map tab (user requested to hide location field) -->
                        <v-tab value="map" :disabled="!transaction.geoLocation" v-if="false && type === TransactionEditPageType.Transaction && !!getMapProvider()">
                            <span>{{ tt('Location on Map') }}</span>
                        </v-tab>
                        <!-- Hidden: Pictures tab (user requested to hide pictures block) -->
                        <v-tab value="pictures" :disabled="mode !== TransactionEditPageMode.Add && mode !== TransactionEditPageMode.Edit && (!transaction.pictures || !transaction.pictures!.length)" v-if="false && type === TransactionEditPageType.Transaction && isTransactionPicturesEnabled()">
                            <span>{{ tt('Pictures') }}</span>
                        </v-tab>
                    </v-tabs>
                </div>

                <v-window class="d-flex flex-grow-1 disable-tab-transition w-100-window-container ms-md-5"
                          v-model="activeTab">
                    <v-window-item value="basicInfo">
                        <v-form class="mt-2">
                            <v-row>
                                <v-col cols="12" v-if="type === TransactionEditPageType.Template && transaction instanceof TransactionTemplate">
                                    <v-text-field
                                        type="text"
                                        persistent-placeholder
                                        :disabled="loading || submitting"
                                        :label="tt('Template Name')"
                                        :placeholder="tt('Template Name')"
                                        v-model="transaction.name"
                                    />
                                </v-col>
                                <v-col cols="12" :md="transaction.type === TransactionType.Transfer ? 6 : 12">
                                    <amount-input class="transaction-edit-amount font-weight-bold"
                                                  :color="sourceAmountColor"
                                                  :currency="sourceAccountCurrency"
                                                  :show-currency="true"
                                                  :readonly="mode === TransactionEditPageMode.View"
                                                  :disabled="loading || submitting"
                                                  :persistent-placeholder="true"
                                                  :hide="transaction.hideAmount"
                                                  :label="sourceAmountTitle"
                                                  :placeholder="tt(sourceAmountName)"
                                                  :enable-formula="mode !== TransactionEditPageMode.View"
                                                  v-model="transaction.sourceAmount"/>
                                </v-col>
                                <v-col cols="12" :md="6" v-if="transaction.type === TransactionType.Transfer">
                                    <amount-input class="transaction-edit-amount font-weight-bold" color="primary"
                                                  :currency="destinationAccountCurrency"
                                                  :show-currency="true"
                                                  :readonly="mode === TransactionEditPageMode.View"
                                                  :disabled="loading || submitting"
                                                  :persistent-placeholder="true"
                                                  :hide="transaction.hideAmount"
                                                  :label="transferInAmountTitle"
                                                  :placeholder="tt('Transfer In Amount')"
                                                  :enable-formula="mode !== TransactionEditPageMode.View"
                                                  v-model="transaction.destinationAmount"/>
                                </v-col>
                                <v-col cols="12" md="12" v-if="transaction.type === TransactionType.Expense">
                                    <v-tooltip :disabled="hasVisibleExpenseCategories" :text="hasVisibleExpenseCategories ? '' : tt('No available expense categories')">
                                        <template v-slot:activator="{ props }">
                                            <div v-bind="props" class="d-block">
                                                <v-select
                                                    item-title="name"
                                                    item-value="id"
                                                    persistent-placeholder
                                                    :readonly="mode === TransactionEditPageMode.View"
                                                    :disabled="loading || submitting || !hasVisibleExpenseCategories"
                                                    :label="tt('Category')"
                                                    :placeholder="tt('Category')"
                                                    :items="allCategories[CategoryType.Expense] || []"
                                                    :no-data-text="tt('No available category')"
                                                    v-model="transaction.expenseCategoryId"
                                                >
                                                    <template #item="{ props: itemProps, item }">
                                                        <v-list-item v-bind="itemProps">
                                                            <template #prepend>
                                                                <ItemIcon class="me-2" icon-type="category"
                                                                          :icon-id="item.raw.icon" :color="item.raw.color"></ItemIcon>
                                                            </template>
                                                        </v-list-item>
                                                    </template>
                                                    <template #selection="{ item }">
                                                        <ItemIcon class="me-2" icon-type="category"
                                                                  :icon-id="item.raw.icon" :color="item.raw.color"></ItemIcon>
                                                        <span>{{ item.raw.name }}</span>
                                                    </template>
                                                </v-select>
                                            </div>
                                        </template>
                                    </v-tooltip>
                                </v-col>
                                <v-col cols="12" md="12" v-if="transaction.type === TransactionType.Income">
                                    <v-tooltip :disabled="hasVisibleIncomeCategories" :text="hasVisibleIncomeCategories ? '' : tt('No available income categories')">
                                        <template v-slot:activator="{ props }">
                                            <div v-bind="props" class="d-block">
                                                <v-select
                                                    item-title="name"
                                                    item-value="id"
                                                    persistent-placeholder
                                                    :readonly="mode === TransactionEditPageMode.View"
                                                    :disabled="loading || submitting || !hasVisibleIncomeCategories"
                                                    :label="tt('Category')"
                                                    :placeholder="tt('Category')"
                                                    :items="allCategories[CategoryType.Income] || []"
                                                    :no-data-text="tt('No available category')"
                                                    v-model="transaction.incomeCategoryId"
                                                >
                                                    <template #item="{ props: itemProps, item }">
                                                        <v-list-item v-bind="itemProps">
                                                            <template #prepend>
                                                                <ItemIcon class="me-2" icon-type="category"
                                                                          :icon-id="item.raw.icon" :color="item.raw.color"></ItemIcon>
                                                            </template>
                                                        </v-list-item>
                                                    </template>
                                                    <template #selection="{ item }">
                                                        <ItemIcon class="me-2" icon-type="category"
                                                                  :icon-id="item.raw.icon" :color="item.raw.color"></ItemIcon>
                                                        <span>{{ item.raw.name }}</span>
                                                    </template>
                                                </v-select>
                                            </div>
                                        </template>
                                    </v-tooltip>
                                </v-col>
                                <v-col cols="12" md="12" v-if="transaction.type === TransactionType.Transfer">
                                    <v-tooltip :disabled="hasVisibleTransferCategories" :text="hasVisibleTransferCategories ? '' : tt('No available transfer categories')">
                                        <template v-slot:activator="{ props }">
                                            <div v-bind="props" class="d-block">
                                                <v-select
                                                    item-title="name"
                                                    item-value="id"
                                                    persistent-placeholder
                                                    :readonly="mode === TransactionEditPageMode.View"
                                                    :disabled="loading || submitting || !hasVisibleTransferCategories"
                                                    :label="tt('Category')"
                                                    :placeholder="tt('Category')"
                                                    :items="allCategories[CategoryType.Transfer] || []"
                                                    :no-data-text="tt('No available category')"
                                                    v-model="transaction.transferCategoryId"
                                                >
                                                    <template #item="{ props: itemProps, item }">
                                                        <v-list-item v-bind="itemProps">
                                                            <template #prepend>
                                                                <ItemIcon class="me-2" icon-type="category"
                                                                          :icon-id="item.raw.icon" :color="item.raw.color"></ItemIcon>
                                                            </template>
                                                        </v-list-item>
                                                    </template>
                                                    <template #selection="{ item }">
                                                        <ItemIcon class="me-2" icon-type="category"
                                                                  :icon-id="item.raw.icon" :color="item.raw.color"></ItemIcon>
                                                        <span>{{ item.raw.name }}</span>
                                                    </template>
                                                </v-select>
                                            </div>
                                        </template>
                                    </v-tooltip>
                                </v-col>
                                <v-col cols="12" md="12" v-if="transaction.type !== TransactionType.ModifyBalance">
                                    <v-autocomplete
                                        item-title="name"
                                        item-value="id"
                                        persistent-placeholder
                                        clearable
                                        auto-select-first
                                        :readonly="mode === TransactionEditPageMode.View"
                                        :disabled="loading || submitting"
                                        :label="tt('Counterparty')"
                                        :placeholder="tt('Counterparty')"
                                        :items="counterpartiesStore.allVisibleCounterparties"
                                        :no-data-text="tt('No available counterparty')"
                                        :model-value="transaction.counterpartyId === '0' ? null : transaction.counterpartyId"
                                        @update:model-value="transaction.counterpartyId = $event || '0'"
                                    >
                                        <template #item="{ props: itemProps, item }">
                                            <v-list-item v-bind="itemProps">
                                                <template #prepend>
                                                    <v-icon class="me-2" :icon="item.raw.type === CounterpartyType.Company ? mdiDomain : mdiAccountOutline"></v-icon>
                                                </template>
                                            </v-list-item>
                                        </template>
                                        <template #selection="{ item }">
                                            <v-icon class="me-2" :icon="item.raw.type === CounterpartyType.Company ? mdiDomain : mdiAccountOutline"></v-icon>
                                            <span>{{ item.raw.name }}</span>
                                        </template>
                                    </v-autocomplete>
                                </v-col>
                                <v-col cols="12" :md="transaction.type === TransactionType.Transfer ? 6 : 12">
                                    <v-tooltip :disabled="!!allVisibleAccounts.length" :text="allVisibleAccounts.length ? '' : tt('No available account')">
                                        <template v-slot:activator="{ props }">
                                            <div v-bind="props" class="d-block">
                                                <two-column-select primary-key-field="id" primary-value-field="category"
                                                                   primary-title-field="name" primary-footer-field="displayBalance"
                                                                   primary-icon-field="icon" primary-icon-type="account"
                                                                   primary-sub-items-field="accounts"
                                                                   :primary-title-i18n="true"
                                                                   secondary-key-field="id" secondary-value-field="id"
                                                                   secondary-title-field="name" secondary-footer-field="displayBalance"
                                                                   secondary-icon-field="icon" secondary-icon-type="account" secondary-color-field="color"
                                                                   :readonly="mode === TransactionEditPageMode.View"
                                                                   :disabled="loading || submitting || !allVisibleAccounts.length || (mode === TransactionEditPageMode.Edit && transaction.type === TransactionType.ModifyBalance)"
                                                                   :enable-filter="true" :filter-placeholder="tt('Find account')" :filter-no-items-text="tt('No available account')"
                                                                   :custom-selection-primary-text="sourceAccountName"
                                                                   :label="tt(sourceAccountTitle)"
                                                                   :placeholder="tt(sourceAccountTitle)"
                                                                   :items="allVisibleCategorizedAccounts"
                                                                   v-model="transaction.sourceAccountId">
                                                </two-column-select>
                                            </div>
                                        </template>
                                    </v-tooltip>
                                </v-col>
                                <v-col cols="12" md="6" v-if="transaction.type === TransactionType.Transfer">
                                    <v-tooltip :disabled="!!allVisibleAccounts.length" :text="allVisibleAccounts.length ? '' : tt('No available account')">
                                        <template v-slot:activator="{ props }">
                                            <div v-bind="props" class="d-block">
                                                <two-column-select primary-key-field="id" primary-value-field="category"
                                                                   primary-title-field="name" primary-footer-field="displayBalance"
                                                                   primary-icon-field="icon" primary-icon-type="account"
                                                                   primary-sub-items-field="accounts"
                                                                   :primary-title-i18n="true"
                                                                   secondary-key-field="id" secondary-value-field="id"
                                                                   secondary-title-field="name" secondary-footer-field="displayBalance"
                                                                   secondary-icon-field="icon" secondary-icon-type="account" secondary-color-field="color"
                                                                   :readonly="mode === TransactionEditPageMode.View"
                                                                   :disabled="loading || submitting || !allVisibleAccounts.length"
                                                                   :enable-filter="true" :filter-placeholder="tt('Find account')" :filter-no-items-text="tt('No available account')"
                                                                   :custom-selection-primary-text="destinationAccountName"
                                                                   :label="tt('Destination Account')"
                                                                   :placeholder="tt('Destination Account')"
                                                                   :items="allVisibleCategorizedAccounts"
                                                                   v-model="transaction.destinationAccountId">
                                                </two-column-select>
                                            </div>
                                        </template>
                                    </v-tooltip>
                                </v-col>
                                <v-col cols="12" md="6" v-if="type === TransactionEditPageType.Transaction">
                                    <date-time-select
                                        :readonly="mode === TransactionEditPageMode.View"
                                        :disabled="loading || submitting || (mode === TransactionEditPageMode.Edit && transaction.type === TransactionType.ModifyBalance)"
                                        :label="tt('Transaction Date')"
                                        :hide-time-picker="true"
                                        :timezone-utc-offset="transaction.utcOffset"
                                        :model-value="transaction.time"
                                        @update:model-value="updateTransactionTime"
                                        @error="onShowDateTimeError" />
                                </v-col>
                                <v-col cols="12" md="6" v-if="type === TransactionEditPageType.Template && transaction instanceof TransactionTemplate && transaction.templateType === TemplateType.Schedule.type">
                                    <schedule-frequency-select
                                        :readonly="mode === TransactionEditPageMode.View"
                                        :disabled="loading || submitting"
                                        :label="tt('Scheduled Transaction Frequency')"
                                        v-model:type="transaction.scheduledFrequencyType"
                                        v-model="transaction.scheduledFrequency" />
                                </v-col>
                                <!-- Hidden: Transaction Timezone (user requested to hide time field) -->
                                <v-col cols="12" md="6" v-if="false && (type === TransactionEditPageType.Transaction || (type === TransactionEditPageType.Template && transaction instanceof TransactionTemplate && (transaction as any).templateType === TemplateType.Schedule.type))">
                                    <v-autocomplete
                                        class="transaction-edit-timezone"
                                        item-title="displayNameWithUtcOffset"
                                        item-value="name"
                                        auto-select-first
                                        persistent-placeholder
                                        :readonly="mode === TransactionEditPageMode.View"
                                        :disabled="loading || submitting || (mode === TransactionEditPageMode.Edit && transaction.type === TransactionType.ModifyBalance)"
                                        :label="tt('Transaction Timezone')"
                                        :placeholder="!transaction.timeZone && transaction.timeZone !== '' ? `(${transactionDisplayTimezone}) ${transactionTimezoneTimeDifference}` : tt('Timezone')"
                                        :items="allTimezones"
                                        :no-data-text="tt('No results')"
                                        :model-value="transaction.timeZone"
                                        @update:model-value="updateTransactionTimezone"
                                    >
                                        <template #selection="{ item }">
                                            <span class="text-truncate" v-if="transaction.timeZone || transaction.timeZone === ''">
                                                {{ item.title }}
                                            </span>
                                        </template>
                                    </v-autocomplete>
                                </v-col>
                                <v-col cols="12" md="6" v-if="type === TransactionEditPageType.Template && transaction instanceof TransactionTemplate && transaction.templateType === TemplateType.Schedule.type">
                                    <date-select
                                        :readonly="mode === TransactionEditPageMode.View"
                                        :disabled="loading || submitting"
                                        :clearable="true"
                                        :label="tt('Start Date')"
                                        :no-data-text="tt('No limit')"
                                        v-model="transaction.scheduledStartDate" />
                                </v-col>
                                <v-col cols="12" md="6" v-if="type === TransactionEditPageType.Template && transaction instanceof TransactionTemplate && transaction.templateType === TemplateType.Schedule.type">
                                    <date-select
                                        :readonly="mode === TransactionEditPageMode.View"
                                        :disabled="loading || submitting"
                                        :clearable="true"
                                        :label="tt('End Date')"
                                        :no-data-text="tt('No limit')"
                                        v-model="transaction.scheduledEndDate" />
                                </v-col>
                                <!-- Hidden: Geographic Location (user requested to hide location field) -->
                                <v-col cols="12" md="12" v-if="false && type === TransactionEditPageType.Transaction">
                                    <v-select
                                        persistent-placeholder
                                        :readonly="mode === TransactionEditPageMode.View"
                                        :disabled="loading || submitting"
                                        :label="tt('Geographic Location')"
                                        v-model="transaction"
                                        v-model:menu="geoMenuState"
                                    >
                                        <template #selection>
                                            <span class="cursor-pointer" v-if="transaction.geoLocation">{{ `(${formatCoordinate(transaction.geoLocation!, coordinateDisplayType)})` }}</span>
                                            <span class="cursor-pointer" v-else-if="!transaction.geoLocation">{{ geoLocationStatusInfo }}</span>
                                        </template>

                                        <template #no-data>
                                            <v-list class="py-0">
                                                <v-list-item v-if="mode !== TransactionEditPageMode.View" @click="updateGeoLocation(true)">{{ tt('Update Geographic Location') }}</v-list-item>
                                                <v-list-item v-if="mode !== TransactionEditPageMode.View" @click="clearGeoLocation">{{ tt('Clear Geographic Location') }}</v-list-item>
                                            </v-list>
                                        </template>
                                    </v-select>
                                </v-col>
                                <v-col cols="12" md="12">
                                    <transaction-tag-auto-complete
                                        :readonly="mode === TransactionEditPageMode.View"
                                        :disabled="loading || submitting"
                                        :show-label="true"
                                        :allow-add-new-tag="true"
                                        v-model="transaction.tagIds"
                                        @tag:saving="onSavingTag"
                                    />
                                </v-col>
                                <v-col cols="12" md="12">
                                    <v-textarea
                                        type="text"
                                        persistent-placeholder
                                        rows="3"
                                        :readonly="mode === TransactionEditPageMode.View"
                                        :disabled="loading || submitting"
                                        :label="tt('Description')"
                                        :placeholder="tt('Your transaction description (optional)')"
                                        v-model="transaction.comment"
                                    />
                                </v-col>
                                <v-col cols="12" md="12" v-if="type === TransactionEditPageType.Transaction && mode === TransactionEditPageMode.Add">
                                    <v-checkbox
                                        density="compact"
                                        :label="tt('Repeatable')"
                                        :disabled="loading || submitting"
                                        v-model="isRepeatable"
                                    />
                                </v-col>
                                <v-col cols="12" md="12" v-if="isRepeatable && type === TransactionEditPageType.Transaction && mode === TransactionEditPageMode.Add">
                                    <schedule-frequency-select
                                        :disabled="loading || submitting"
                                        :label="tt('Scheduled Transaction Frequency')"
                                        v-model:type="repeatFrequencyType"
                                        v-model="repeatFrequency" />
                                </v-col>
                            </v-row>
                        </v-form>
                    </v-window-item>
                    <v-window-item value="map">
                        <v-row>
                            <v-col cols="12" md="12">
                                <map-view ref="map" map-class="transaction-edit-map-view"
                                          :enable-zoom-control="true" :geo-location="transaction.geoLocation"
                                          @click="updateSpecifiedGeoLocation">
                                    <template #error-title="{ mapSupported, mapDependencyLoaded }">
                                        <span class="text-subtitle-1" v-if="!mapSupported"><b>{{ tt('Unsupported Map Provider') }}</b></span>
                                        <span class="text-subtitle-1" v-else-if="!mapDependencyLoaded"><b>{{ tt('Cannot Initialize Map') }}</b></span>
                                    </template>
                                    <template #error-content>
                                        <p class="text-body-1">
                                            {{ tt('Please refresh the page and try again. If the error persists, ensure that the server\'s map settings are correctly configured.') }}
                                        </p>
                                    </template>
                                </map-view>
                            </v-col>
                        </v-row>
                    </v-window-item>
                    <!-- Hidden: Pictures window-item (user requested to hide pictures block) -->
                    <v-window-item value="pictures" v-if="false">
                        <v-row class="transaction-pictures align-content-start" :class="{ 'readonly': submitting || uploadingPicture || removingPictureId }">
                            <v-col :key="picIdx" cols="6" md="3" v-for="(pictureInfo, picIdx) in transaction.pictures">
                                <v-avatar rounded="lg" variant="tonal" size="160"
                                          class="cursor-pointer transaction-picture"
                                          color="rgba(0,0,0,0)" @click="viewOrRemovePicture(pictureInfo)">
                                    <v-img :src="getTransactionPictureUrl(pictureInfo)">
                                        <template #placeholder>
                                            <div class="d-flex align-center justify-center fill-height bg-light-primary">
                                                <v-progress-circular color="grey-500" indeterminate size="48"></v-progress-circular>
                                            </div>
                                        </template>
                                        <template #error>
                                            <div class="d-flex align-center justify-center fill-height bg-light-primary">
                                                <span class="text-body-1">{{ tt('Failed to load image, please check whether the config "domain" and "root_url" are set correctly.') }}</span>
                                            </div>
                                        </template>
                                    </v-img>
                                    <div class="picture-control-icon" :class="{ 'show-control-icon': pictureInfo.pictureId === removingPictureId }">
                                        <v-icon size="64" :icon="mdiTrashCanOutline" v-if="(mode === TransactionEditPageMode.Add || mode === TransactionEditPageMode.Edit) && pictureInfo.pictureId !== removingPictureId"/>
                                        <v-progress-circular color="grey-500" indeterminate size="48" v-if="(mode === TransactionEditPageMode.Add || mode === TransactionEditPageMode.Edit) && pictureInfo.pictureId === removingPictureId"></v-progress-circular>
                                        <v-icon size="64" :icon="mdiFullscreen" v-if="mode !== TransactionEditPageMode.Add && mode !== TransactionEditPageMode.Edit"/>
                                    </div>
                                </v-avatar>
                            </v-col>
                            <v-col cols="6" md="3" v-if="canAddTransactionPicture">
                                <v-avatar rounded="lg" variant="tonal" size="160"
                                          class="transaction-picture transaction-picture-add"
                                          :class="{ 'enabled': !submitting, 'cursor-pointer': !submitting }"
                                          color="rgba(0,0,0,0)" @click="showOpenPictureDialog">
                                    <v-tooltip activator="parent" v-if="!submitting">{{ tt('Add Picture') }}</v-tooltip>
                                    <v-icon class="transaction-picture-add-icon" size="56" :icon="mdiImagePlusOutline" v-if="!uploadingPicture"/>
                                    <v-progress-circular color="grey-500" indeterminate size="48" v-if="uploadingPicture"></v-progress-circular>
                                </v-avatar>
                            </v-col>
                        </v-row>
                    </v-window-item>
                </v-window>
            </v-card-text>
            <v-card-text>
                <div class="w-100 d-flex justify-center flex-wrap mt-sm-1 mt-md-2 gap-4">
                    <v-tooltip :disabled="!inputIsEmpty" :text="inputEmptyProblemMessage ? tt(inputEmptyProblemMessage) : ''">
                        <template v-slot:activator="{ props }">
                            <div v-bind="props" class="d-inline-block">
                                <v-btn :disabled="inputIsEmpty || loading || submitting"
                                       v-if="mode !== TransactionEditPageMode.View" @click="save">
                                    {{ tt(saveButtonTitle) }}
                                    <v-progress-circular indeterminate size="22" class="ms-2" v-if="submitting"></v-progress-circular>
                                </v-btn>
                            </div>
                        </template>
                    </v-tooltip>
                    <v-btn-group variant="tonal" density="comfortable"
                                 v-if="mode === TransactionEditPageMode.View && transaction.type !== TransactionType.ModifyBalance">
                        <v-btn :disabled="loading || submitting"
                               @click="duplicate(false, false)">{{ tt('Duplicate') }}</v-btn>
                        <v-btn density="compact" :disabled="loading || submitting" :icon="true">
                            <v-icon :icon="mdiMenuDown" size="24" />
                            <v-menu activator="parent">
                                <v-list>
                                    <v-list-item :title="tt('Duplicate (With Time)')"
                                                 @click="duplicate(true, false)"></v-list-item>
                                    <v-list-item :title="tt('Duplicate (With Geographic Location)')"
                                                 @click="duplicate(false, true)"
                                                 v-if="transaction.geoLocation"></v-list-item>
                                    <v-list-item :title="tt('Duplicate (With Time and Geographic Location)')"
                                                 @click="duplicate(true, true)"
                                                 v-if="transaction.geoLocation"></v-list-item>
                                </v-list>
                            </v-menu>
                        </v-btn>
                    </v-btn-group>
                    <v-btn color="warning" variant="tonal" :disabled="loading || submitting"
                           v-if="mode === TransactionEditPageMode.View && originalTransactionEditable"
                           @click="edit">{{ tt('Edit') }}</v-btn>
                    <v-btn color="info" variant="tonal" :disabled="loading || submitting"
                           v-if="mode === TransactionEditPageMode.View && originalTransactionEditable && transaction.type !== TransactionType.Transfer && transaction.type !== TransactionType.ModifyBalance"
                           @click="openSplitDialog">{{ tt('Split') }}</v-btn>
                    <v-btn color="success" variant="tonal" :disabled="loading || submitting"
                           v-if="mode === TransactionEditPageMode.View && transaction.planned"
                           @click="confirmPlanned">
                        {{ tt('Confirm') }}
                        <v-progress-circular indeterminate size="22" class="ms-2" v-if="submitting"></v-progress-circular>
                    </v-btn>
                    <v-btn color="error" variant="tonal" :disabled="loading || submitting"
                           v-if="mode === TransactionEditPageMode.View && originalTransactionEditable" @click="remove">
                        {{ tt('Delete') }}
                        <v-progress-circular indeterminate size="22" class="ms-2" v-if="submitting"></v-progress-circular>
                    </v-btn>
                    <v-btn color="secondary" variant="tonal" :disabled="loading || submitting"
                           @click="cancel">{{ tt(cancelButtonTitle) }}</v-btn>
                </div>
            </v-card-text>
        </v-card>
    </v-dialog>

    <confirm-dialog ref="confirmDialog"/>
    <snack-bar ref="snackbar" />

    <v-dialog persistent min-width="360" width="auto" v-model="showDeletePlannedDialog">
        <v-card>
            <v-toolbar color="error">
                <v-toolbar-title>{{ tt('Delete Planned Transaction') }}</v-toolbar-title>
            </v-toolbar>
            <v-card-text class="pa-4 pb-6">{{ tt('This is a planned transaction. What do you want to delete?') }}</v-card-text>
            <v-card-actions class="px-4 pb-4 d-flex flex-wrap justify-end ga-2">
                <v-btn color="gray" @click="showDeletePlannedDialog = false">{{ tt('Cancel') }}</v-btn>
                <v-btn color="error" variant="tonal" @click="showDeletePlannedDialog = false; doDeleteOne()">{{ tt('Delete Only This') }}</v-btn>
                <v-btn color="error" @click="showDeletePlannedDialog = false; doDeleteAllFuture()">{{ tt('Delete All Future') }}</v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>

    <v-dialog persistent min-width="500" max-width="700" v-model="showSplitDialog">
        <v-card>
            <v-toolbar color="info">
                <v-toolbar-title>{{ tt('Split Transaction') }}</v-toolbar-title>
            </v-toolbar>
            <v-card-text class="pa-4">
                <div class="text-subtitle-1 mb-4">
                    {{ tt('Total Amount') }}: <strong>{{ formatAmountToLocalizedNumerals(transaction.sourceAmount / 100, sourceAccountCurrency) }}</strong>
                </div>

                <div v-for="(part, idx) in splitParts" :key="idx" class="d-flex align-center ga-3 mb-3">
                    <span class="text-body-2 text-no-wrap" style="min-width: 60px">{{ tt('Part') }} {{ idx + 1 }}:</span>
                    <amount-input style="flex: 1; min-width: 150px"
                                  :currency="sourceAccountCurrency"
                                  :show-currency="true"
                                  :label="tt('Amount')"
                                  v-model="part.amount" />
                    <v-select style="flex: 1.5; min-width: 200px"
                              item-title="name" item-value="id" density="compact"
                              :label="tt('Category')"
                              :items="splitCategoryItems"
                              v-model="part.categoryId">
                        <template #item="{ props: itemProps, item }">
                            <v-list-item v-bind="itemProps">
                                <template #prepend>
                                    <ItemIcon class="me-2" icon-type="category"
                                              :icon-id="item.raw.icon" :color="item.raw.color"></ItemIcon>
                                </template>
                            </v-list-item>
                        </template>
                        <template #selection="{ item }">
                            <ItemIcon class="me-2" icon-type="category"
                                      :icon-id="item.raw.icon" :color="item.raw.color"></ItemIcon>
                            <span>{{ item.raw.name }}</span>
                        </template>
                    </v-select>
                    <v-btn icon size="small" variant="text" color="error"
                           v-if="splitParts.length > 2"
                           @click="removeSplitPart(idx)">
                        <v-icon :icon="mdiClose" />
                    </v-btn>
                    <div v-else style="width: 28px"></div>
                </div>

                <div class="d-flex justify-center mb-3">
                    <v-btn variant="tonal" size="small" @click="addSplitPart">
                        <v-icon :icon="mdiPlus" class="me-1" />
                        {{ tt('Add Part') }}
                    </v-btn>
                </div>

                <v-divider class="mb-3" />

                <div class="d-flex justify-space-between align-center">
                    <span class="text-body-2">{{ tt('Remainder') }}:</span>
                    <span :class="splitRemainder === 0 ? 'text-success' : 'text-error'" class="text-subtitle-1 font-weight-bold">
                        {{ formatAmountToLocalizedNumerals(splitRemainder / 100, sourceAccountCurrency) }}
                        <v-icon v-if="splitRemainder === 0" :icon="mdiCheck" color="success" size="small" />
                    </span>
                </div>
            </v-card-text>
            <v-card-actions class="px-4 pb-4">
                <v-spacer />
                <v-btn color="gray" @click="showSplitDialog = false">{{ tt('Cancel') }}</v-btn>
                <v-btn color="info" :disabled="!canPerformSplit || splittingTransaction"
                       @click="performSplit">
                    {{ tt('Split') }}
                    <v-progress-circular indeterminate size="22" class="ms-2" v-if="splittingTransaction"></v-progress-circular>
                </v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>
    <input ref="pictureInput" type="file" style="display: none" :accept="SUPPORTED_IMAGE_EXTENSIONS" @change="uploadPicture($event)" />
</template>

<script setup lang="ts">
import MapView from '@/components/common/MapView.vue';
import ConfirmDialog from '@/components/desktop/ConfirmDialog.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';

import { ref, computed, useTemplateRef, watch, nextTick } from 'vue';

import { useI18n } from '@/locales/helpers.ts';
import {
    TransactionEditPageMode,
    TransactionEditPageType,
    GeoLocationStatus,
    useTransactionEditPageBase
} from '@/views/base/transactions/TransactionEditPageBase.ts';

import { useSettingsStore } from '@/stores/setting.ts';
import { useUserStore } from '@/stores/user.ts';
import { useAccountsStore } from '@/stores/account.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useTransactionTagsStore } from '@/stores/transactionTag.ts';
import { useTransactionsStore } from '@/stores/transaction.ts';
import { useTransactionTemplatesStore } from '@/stores/transactionTemplate.ts';
import { useCounterpartiesStore } from '@/stores/counterparty.ts';

import type { Coordinate } from '@/core/coordinate.ts';
import { CategoryType } from '@/core/category.ts';
import { TransactionType, TransactionEditScopeType } from '@/core/transaction.ts';
import { TemplateType, ScheduledTemplateFrequencyType } from '@/core/template.ts';
import { KnownErrorCode } from '@/consts/api.ts';
import { SUPPORTED_IMAGE_EXTENSIONS } from '@/consts/file.ts';

import { TransactionTemplate } from '@/models/transaction_template.ts';
import type { TransactionPictureInfoBasicResponse } from '@/models/transaction_picture_info.ts';
import { Transaction } from '@/models/transaction.ts';

import {
    getTimezoneOffsetMinutes,
    getCurrentUnixTime
} from '@/lib/datetime.ts';
import { formatCoordinate } from '@/lib/coordinate.ts';
import { generateRandomUUID } from '@/lib/misc.ts';
import { type SetTransactionOptions, setTransactionModelByTransaction } from '@/lib/transaction.ts';
import {
    isTransactionPicturesEnabled,
    getMapProvider
} from '@/lib/server_settings.ts';
import {
    isSupportGetGeoLocationByClick
} from '@/lib/map/index.ts';
import services from '@/lib/services.ts';
import logger from '@/lib/logger.ts';

import { CounterpartyType } from '@/models/counterparty.ts';

import {
    mdiDotsVertical,
    mdiEyeOffOutline,
    mdiEyeOutline,
    mdiSwapHorizontal,
    mdiMapMarkerOutline,
    mdiCheck,
    mdiMenuDown,
    mdiImagePlusOutline,
    mdiTrashCanOutline,
    mdiFullscreen,
    mdiDomain,
    mdiAccountOutline,
    mdiClose,
    mdiPlus
} from '@mdi/js';

export interface TransactionEditOptions extends SetTransactionOptions {
    id?: string;
    templateType?: number;
    template?: TransactionTemplate;
    currentTransaction?: Transaction;
    currentTemplate?: TransactionTemplate;
    noTransactionDraft?: boolean;
}

interface TransactionEditResponse {
    message: string;
}

type MapViewType = InstanceType<typeof MapView>;
type ConfirmDialogType = InstanceType<typeof ConfirmDialog>;
type SnackBarType = InstanceType<typeof SnackBar>;

const props = defineProps<{
    type: TransactionEditPageType;
    persistent?: boolean;
    show?: boolean;
}>();

const { tt, formatAmountToLocalizedNumerals } = useI18n();

const {
    mode,
    isSupportGeoLocation,
    editId,
    addByTemplateId,
    duplicateFromId,
    clientSessionId,
    loading,
    submitting,
    uploadingPicture,
    geoLocationStatus,
    setGeoLocationByClickMap,
    transaction,
    defaultCurrency,
    defaultAccountId,
    coordinateDisplayType,
    allTimezones,
    allVisibleAccounts,
    allAccountsMap,
    allVisibleCategorizedAccounts,
    allCategories,
    allCategoriesMap,
    allTagsMap,
    firstVisibleAccountId,
    hasVisibleExpenseCategories,
    hasVisibleIncomeCategories,
    hasVisibleTransferCategories,
    canAddTransactionPicture,
    title,
    saveButtonTitle,
    cancelButtonTitle,
    sourceAmountName,
    sourceAmountTitle,
    sourceAccountTitle,
    transferInAmountTitle,
    sourceAccountName,
    destinationAccountName,
    sourceAccountCurrency,
    destinationAccountCurrency,
    transactionDisplayTimezone,
    transactionTimezoneTimeDifference,
    geoLocationStatusInfo,
    inputEmptyProblemMessage,
    inputIsEmpty,
    createNewTransactionModel,
    updateTransactionTime,
    updateTransactionTimezone,
    swapTransactionData,
    getTransactionPictureUrl
} = useTransactionEditPageBase(props.type);

const settingsStore = useSettingsStore();
const userStore = useUserStore();
const accountsStore = useAccountsStore();
const transactionCategoriesStore = useTransactionCategoriesStore();
const transactionTagsStore = useTransactionTagsStore();
const transactionsStore = useTransactionsStore();
const transactionTemplatesStore = useTransactionTemplatesStore();
const counterpartiesStore = useCounterpartiesStore();

const map = useTemplateRef<MapViewType>('map');
const confirmDialog = useTemplateRef<ConfirmDialogType>('confirmDialog');
const snackbar = useTemplateRef<SnackBarType>('snackbar');
const pictureInput = useTemplateRef<HTMLInputElement>('pictureInput');

const showState = ref<boolean>(false);
const activeTab = ref<string>('basicInfo');
const originalTransactionEditable = ref<boolean>(false);
const isRepeatable = ref<boolean>(false);
const repeatFrequencyType = ref<number>(ScheduledTemplateFrequencyType.Monthly.type);
const repeatFrequency = ref<string>('1');
const noTransactionDraft = ref<boolean>(false);
const geoMenuState = ref<boolean>(false);
const removingPictureId = ref<string>('');
const showDeletePlannedDialog = ref<boolean>(false);
const showSplitDialog = ref<boolean>(false);
const splitParts = ref<{ amount: number; categoryId: string }[]>([]);
const splittingTransaction = ref<boolean>(false);

const initAmount = ref<number | undefined>(undefined);
const initCategoryId = ref<string | undefined>(undefined);
const initAccountId = ref<string | undefined>(undefined);
const initTagIds = ref<string | undefined>(undefined);

let resolveFunc: ((response?: TransactionEditResponse) => void) | null = null;
let rejectFunc: ((reason?: unknown) => void) | null = null;

const sourceAmountColor = computed<string | undefined>(() => {
    if (transaction.value.type === TransactionType.Expense) {
        return 'expense';
    } else if (transaction.value.type === TransactionType.Income) {
        return 'income';
    } else if (transaction.value.type === TransactionType.Transfer) {
        return 'primary';
    }

    return undefined;
});



const isTransactionModified = computed<boolean>(() => {
    if (mode.value === TransactionEditPageMode.Add) {
        return transactionsStore.isTransactionDraftModified(transaction.value, initAmount.value, initCategoryId.value, initAccountId.value, initTagIds.value, firstVisibleAccountId.value);
    } else if (mode.value === TransactionEditPageMode.Edit) {
        return true;
    } else {
        return false;
    }
});

function setTransaction(newTransaction: Transaction | null, options: SetTransactionOptions, setContextData: boolean): void {
    setTransactionModelByTransaction(
        transaction.value,
        newTransaction,
        allCategories.value,
        allCategoriesMap.value,
        allVisibleAccounts.value,
        allAccountsMap.value,
        allTagsMap.value,
        defaultAccountId.value,
        {
            time: options.time,
            type: options.type,
            categoryId: options.categoryId,
            accountId: options.accountId,
            destinationAccountId: options.destinationAccountId,
            amount: options.amount,
            destinationAmount: options.destinationAmount,
            tagIds: options.tagIds,
            comment: options.comment
        },
        setContextData
    );
}

function open(options: TransactionEditOptions): Promise<TransactionEditResponse | undefined> {
    addByTemplateId.value = null;
    duplicateFromId.value = null;
    showState.value = true;
    activeTab.value = 'basicInfo';
    loading.value = true;
    submitting.value = false;
    geoLocationStatus.value = null;
    setGeoLocationByClickMap.value = false;
    originalTransactionEditable.value = false;
    noTransactionDraft.value = options.noTransactionDraft || false;

    initAmount.value = options.amount;
    initCategoryId.value = options.categoryId;
    initAccountId.value = options.accountId;
    initTagIds.value = options.tagIds;

    const newTransaction = createNewTransactionModel(options.type);
    setTransaction(newTransaction, options, true);

    const promises: Promise<unknown>[] = [
        accountsStore.loadAllAccounts({ force: false }),
        transactionCategoriesStore.loadAllCategories({ force: false }),
        transactionTagsStore.loadAllTags({ force: false })
    ];

    counterpartiesStore.loadAllCounterparties({ force: false }).catch(() => {
        // counterparties are optional, ignore load errors
    });

    if (props.type === TransactionEditPageType.Transaction) {
        if (options && options.id) {
            if (options.currentTransaction) {
                setTransaction(options.currentTransaction, options, true);
            }

            mode.value = TransactionEditPageMode.View;
            editId.value = options.id;

            promises.push(transactionsStore.getTransaction({ transactionId: editId.value }));
        } else {
            mode.value = TransactionEditPageMode.Add;
            editId.value = null;

            if (options.template) {
                setTransaction(options.template, options, false);
                addByTemplateId.value = options.template.id;
            } else if (!options.noTransactionDraft && (settingsStore.appSettings.autoSaveTransactionDraft === 'enabled' || settingsStore.appSettings.autoSaveTransactionDraft === 'confirmation') && transactionsStore.transactionDraft) {
                setTransaction(Transaction.ofDraft(transactionsStore.transactionDraft), options, false);
            }

            if (settingsStore.appSettings.autoGetCurrentGeoLocation
                && !geoLocationStatus.value && !transaction.value.geoLocation) {
                updateGeoLocation(false);
            }
        }
    } else if (props.type === TransactionEditPageType.Template) {
        const template = TransactionTemplate.createNewTransactionTemplate(transaction.value);
        template.name = '';

        if (options && options.templateType) {
            template.templateType = options.templateType;
        }

        if (template.templateType === TemplateType.Schedule.type) {
            template.scheduledFrequencyType = ScheduledTemplateFrequencyType.Disabled.type;
            template.scheduledFrequency = '';
        }

        transaction.value = template;

        if (options && options.id) {
            if (options.currentTemplate) {
                setTransaction(options.currentTemplate, options, false);
                (transaction.value as TransactionTemplate).fillFrom(options.currentTemplate);
            }

            mode.value = TransactionEditPageMode.Edit;
            editId.value = options.id;
            transaction.value.id = options.id;

            promises.push(transactionTemplatesStore.getTemplate({ templateId: editId.value }));
        } else {
            mode.value = TransactionEditPageMode.Add;
            editId.value = null;
            transaction.value.id = '';
        }
    }

    if (options.type &&
        options.type >= TransactionType.Income &&
        options.type <= TransactionType.Transfer) {
        transaction.value.type = options.type;
    }

    if (mode.value === TransactionEditPageMode.Add) {
        clientSessionId.value = generateRandomUUID();
    }

    Promise.all(promises).then(function (responses) {
        if (editId.value && !responses[3]) {
            if (rejectFunc) {
                if (props.type === TransactionEditPageType.Transaction) {
                    rejectFunc('Unable to retrieve transaction');
                } else if (props.type === TransactionEditPageType.Template) {
                    rejectFunc('Unable to retrieve template');
                }
            }

            return;
        }

        if (props.type === TransactionEditPageType.Transaction && options && options.id && responses[3] && responses[3] instanceof Transaction) {
            const transaction: Transaction = responses[3];
            setTransaction(transaction, options, true);
            originalTransactionEditable.value = transaction.editable;
        } else if (props.type === TransactionEditPageType.Template && options && options.id && responses[3] && responses[3] instanceof TransactionTemplate) {
            const template: TransactionTemplate = responses[3];
            setTransaction(template, options, false);

            if (!(transaction.value instanceof TransactionTemplate)) {
                transaction.value = TransactionTemplate.createNewTransactionTemplate(transaction.value);
            }

            (transaction.value as TransactionTemplate).fillFrom(template);
        } else {
            setTransaction(null, options, true);
        }

        loading.value = false;
    }).catch(error => {
        logger.error('failed to load essential data for editing transaction', error);

        loading.value = false;
        showState.value = false;

        if (!error.processed) {
            if (rejectFunc) {
                rejectFunc(error);
            }
        }
    });

    return new Promise((resolve, reject) => {
        resolveFunc = resolve;
        rejectFunc = reject;
    });
}

function save(): void {
    const problemMessage = inputEmptyProblemMessage.value;

    if (problemMessage) {
        snackbar.value?.showMessage(problemMessage);
        return;
    }

    if (props.type === TransactionEditPageType.Transaction && (mode.value === TransactionEditPageMode.Add || mode.value === TransactionEditPageMode.Edit)) {
        const doSubmit = function () {
            submitting.value = true;

            transactionsStore.saveTransaction({
                transaction: transaction.value as Transaction,
                defaultCurrency: defaultCurrency.value,
                isEdit: mode.value === TransactionEditPageMode.Edit,
                clientSessionId: clientSessionId.value,
                repeatOptions: isRepeatable.value && mode.value === TransactionEditPageMode.Add ? {
                    repeatable: true,
                    repeatFrequencyType: repeatFrequencyType.value,
                    repeatFrequency: repeatFrequency.value
                } : undefined
            }).then(() => {
                submitting.value = false;

                const afterSave = () => {
                    if (resolveFunc) {
                        if (mode.value === TransactionEditPageMode.Add) {
                            resolveFunc({
                                message: 'You have added a new transaction'
                            });
                        } else if (mode.value === TransactionEditPageMode.Edit) {
                            resolveFunc({
                                message: 'You have saved this transaction'
                            });
                        }
                    }

                    if (mode.value === TransactionEditPageMode.Add && !noTransactionDraft.value && !addByTemplateId.value && !duplicateFromId.value) {
                        transactionsStore.clearTransactionDraft();
                    }

                    showState.value = false;
                };

                // If editing a planned transaction, ask about modifying all future
                if (mode.value === TransactionEditPageMode.Edit && transaction.value.planned) {
                    confirmDialog.value?.open(tt('Do you want to apply these changes to all future planned transactions?')).then(() => {
                        submitting.value = true;

                        services.modifyAllFuturePlannedTransactions({
                            id: transaction.value.id,
                            sourceAmount: transaction.value.sourceAmount,
                            categoryId: transaction.value.categoryId || '0',
                            sourceAccountId: transaction.value.sourceAccountId || '0',
                            destinationAccountId: transaction.value.destinationAccountId || '0',
                            destinationAmount: transaction.value.destinationAmount,
                            hideAmount: transaction.value.hideAmount,
                            counterpartyId: transaction.value.counterpartyId || '0',
                            comment: transaction.value.comment
                        }).then(() => {
                            submitting.value = false;
                            afterSave();
                        }).catch(error => {
                            submitting.value = false;

                            if (!error.processed) {
                                snackbar.value?.showError(error);
                            }

                            afterSave();
                        });
                    }).catch(() => {
                        // User chose not to modify all future — just save this one
                        afterSave();
                    });
                } else {
                    afterSave();
                }
            }).catch(error => {
                submitting.value = false;

                if (error.error && (error.error.errorCode === KnownErrorCode.TransactionCannotCreateInThisTime || error.error.errorCode === KnownErrorCode.TransactionCannotModifyInThisTime)) {
                    confirmDialog.value?.open('You have set this time range to prevent editing transactions. Would you like to change the editable transaction range to All?').then(() => {
                        submitting.value = true;

                        userStore.updateUserTransactionEditScope({
                            transactionEditScope: TransactionEditScopeType.All.type
                        }).then(() => {
                            submitting.value = false;

                            snackbar.value?.showMessage('Your editable transaction range has been set to All');
                        }).catch(error => {
                            submitting.value = false;

                            if (!error.processed) {
                                snackbar.value?.showError(error);
                            }
                        });
                    });
                } else if (!error.processed) {
                    snackbar.value?.showError(error);
                }
            });
        };

        if (transaction.value.sourceAmount === 0) {
            confirmDialog.value?.open('Are you sure you want to save this transaction with a zero amount?').then(() => {
                doSubmit();
            });
        } else {
            doSubmit();
        }
    } else if (props.type === TransactionEditPageType.Template && (mode.value === TransactionEditPageMode.Add || mode.value === TransactionEditPageMode.Edit)) {
        submitting.value = true;

        transactionTemplatesStore.saveTemplateContent({
            template: transaction.value as TransactionTemplate,
            isEdit: mode.value === TransactionEditPageMode.Edit,
            clientSessionId: clientSessionId.value
        }).then(() => {
            submitting.value = false;

            if (resolveFunc) {
                if (mode.value === TransactionEditPageMode.Add) {
                    resolveFunc({
                        message: 'You have added a new template'
                    });
                } else if (mode.value === TransactionEditPageMode.Edit) {
                    resolveFunc({
                        message: 'You have saved this template'
                    });
                }
            }

            showState.value = false;
        }).catch(error => {
            submitting.value = false;

            if (!error.processed) {
                snackbar.value?.showError(error);
            }
        });
    }
}

function duplicate(withTime?: boolean, withGeoLocation?: boolean): void {
    if (props.type !== TransactionEditPageType.Transaction || mode.value !== TransactionEditPageMode.View) {
        return;
    }

    editId.value = null;
    duplicateFromId.value = transaction.value.id;
    clientSessionId.value = generateRandomUUID();
    activeTab.value = 'basicInfo';
    transaction.value.id = '';

    if (!withTime) {
        transaction.value.time = getCurrentUnixTime();
        transaction.value.timeZone = settingsStore.appSettings.timeZone;
        transaction.value.utcOffset = getTimezoneOffsetMinutes(transaction.value.time, transaction.value.timeZone);
    }

    if (!withGeoLocation) {
        transaction.value.removeGeoLocation();
    }

    transaction.value.clearPictures();
    mode.value = TransactionEditPageMode.Add;
}

function edit(): void {
    if (props.type !== TransactionEditPageType.Transaction || mode.value !== TransactionEditPageMode.View) {
        return;
    }

    mode.value = TransactionEditPageMode.Edit;
}

function doDeleteOne(): void {
    submitting.value = true;

    transactionsStore.deleteTransaction({
        transaction: transaction.value as Transaction,
        defaultCurrency: defaultCurrency.value
    }).then(() => {
        if (resolveFunc) {
            resolveFunc();
        }

        submitting.value = false;
        showState.value = false;
    }).catch(error => {
        submitting.value = false;

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function doDeleteAllFuture(): void {
    submitting.value = true;

    services.deleteAllFuturePlannedTransactions({
        id: transaction.value.id
    }).then(() => {
        if (resolveFunc) {
            resolveFunc();
        }

        submitting.value = false;
        showState.value = false;
    }).catch(error => {
        submitting.value = false;

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function remove(): void {
    if (props.type !== TransactionEditPageType.Transaction || mode.value !== TransactionEditPageMode.View) {
        return;
    }

    if (transaction.value.planned) {
        showDeletePlannedDialog.value = true;
    } else {
        confirmDialog.value?.open('Are you sure you want to delete this transaction?').then(() => {
            doDeleteOne();
        });
    }
}

// Split transaction logic
const splitCategoryItems = computed(() => {
    if (transaction.value.type === TransactionType.Expense) {
        return allCategories.value[CategoryType.Expense] || [];
    } else if (transaction.value.type === TransactionType.Income) {
        return allCategories.value[CategoryType.Income] || [];
    }
    return [];
});

const splitTotalAmount = computed(() => splitParts.value.reduce((sum, p) => sum + p.amount, 0));
const splitRemainder = computed(() => transaction.value.sourceAmount - splitTotalAmount.value);
const canPerformSplit = computed(() =>
    splitRemainder.value === 0
    && splitParts.value.length >= 2
    && splitParts.value.every(p => p.amount > 0 && !!p.categoryId)
);

function openSplitDialog(): void {
    splitParts.value = [
        { amount: transaction.value.sourceAmount, categoryId: transaction.value.categoryId || '' }
    ];
    splittingTransaction.value = false;
    showSplitDialog.value = true;
}

function addSplitPart(): void {
    splitParts.value.push({ amount: 0, categoryId: '' });
}

function removeSplitPart(index: number): void {
    if (splitParts.value.length > 2) {
        splitParts.value.splice(index, 1);
    }
}

async function performSplit(): Promise<void> {
    if (!canPerformSplit.value) return;
    splittingTransaction.value = true;

    try {
        // Step 1: Modify original — set amount and category of first part
        const firstPart = splitParts.value[0]!;
        const origReq = (transaction.value as Transaction).toModifyRequest();
        await services.modifyTransaction({
            ...origReq,
            sourceAmount: firstPart.amount,
            categoryId: firstPart.categoryId
        } as any);

        // Step 2: Create new transactions for parts 2..N
        for (let i = 1; i < splitParts.value.length; i++) {
            const part = splitParts.value[i]!;
            await services.addTransaction({
                type: transaction.value.type,
                categoryId: part.categoryId,
                time: transaction.value.time,
                utcOffset: transaction.value.utcOffset,
                sourceAccountId: transaction.value.sourceAccountId,
                destinationAccountId: '0',
                sourceAmount: part.amount,
                destinationAmount: 0,
                hideAmount: transaction.value.hideAmount,
                tagIds: transaction.value.tagIds || [],
                counterpartyId: transaction.value.counterpartyId || '0',
                pictureIds: [],
                comment: transaction.value.comment,
                clientSessionId: generateRandomUUID()
            });
        }

        // Invalidate caches
        transactionsStore.updateTransactionListInvalidState(true);

        splittingTransaction.value = false;
        showSplitDialog.value = false;

        if (resolveFunc) {
            resolveFunc({ message: 'Transaction has been split' });
        }
        showState.value = false;
    } catch (error: any) {
        splittingTransaction.value = false;

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    }
}

async function confirmPlanned(): Promise<void> {
    if (!transaction.value || !transaction.value.id) {
        return;
    }

    const result = await confirmDialog.value?.open(tt('Are you sure you want to confirm this planned transaction? The transaction date will be set to today.'));

    if (!result) {
        return;
    }

    submitting.value = true;

    try {
        const response = await services.confirmPlannedTransaction({ id: transaction.value.id });

        if (response && response.data && response.data.success) {
            snackbar.value?.showMessage(tt('Transaction confirmed successfully'));
            showState.value = false;

            if (resolveFunc) {
                resolveFunc({ message: tt('Transaction confirmed successfully') });
            }
        }
    } catch (error) {
        snackbar.value?.showMessage(tt('Unable to confirm planned transaction'));
        logger.error('failed to confirm planned transaction', error);
    } finally {
        submitting.value = false;
    }
}

function cancel(): void {
    const doClose = function () {
        if (rejectFunc) {
            rejectFunc();
        }

        showState.value = false;
    };

    if (props.type !== TransactionEditPageType.Transaction || mode.value !== TransactionEditPageMode.Add || noTransactionDraft.value || addByTemplateId.value || duplicateFromId.value) {
        doClose();
        return;
    }

    if (settingsStore.appSettings.autoSaveTransactionDraft === 'confirmation') {
        if (transactionsStore.isTransactionDraftModified(transaction.value, initAmount.value, initCategoryId.value, initAccountId.value, initTagIds.value, firstVisibleAccountId.value)) {
            confirmDialog.value?.open('Do you want to save this transaction draft?').then(() => {
                transactionsStore.saveTransactionDraft(transaction.value, initAmount.value, initCategoryId.value, initAccountId.value, initTagIds.value, firstVisibleAccountId.value);
                doClose();
            }).catch(() => {
                transactionsStore.clearTransactionDraft();
                doClose();
            });
        } else {
            transactionsStore.clearTransactionDraft();
            doClose();
        }
    } else if (settingsStore.appSettings.autoSaveTransactionDraft === 'enabled') {
        transactionsStore.saveTransactionDraft(transaction.value, initAmount.value, initCategoryId.value, initAccountId.value, initTagIds.value, firstVisibleAccountId.value);
        doClose();
    } else {
        doClose();
    }
}

function updateGeoLocation(forceUpdate: boolean): void {
    geoMenuState.value = false;

    if (!isSupportGeoLocation) {
        logger.warn('this browser does not support geo location');

        if (forceUpdate) {
            snackbar.value?.showMessage('Unable to retrieve current position');
        }
        return;
    }

    navigator.geolocation.getCurrentPosition(function (position) {
        if (!position || !position.coords) {
            logger.error('current position is null');
            geoLocationStatus.value = GeoLocationStatus.Error;

            if (forceUpdate) {
                snackbar.value?.showMessage('Unable to retrieve current position');
            }

            return;
        }

        geoLocationStatus.value = GeoLocationStatus.Success;

        transaction.value.setLatitudeAndLongitude(position.coords.latitude, position.coords.longitude);
    }, function (err) {
        logger.error('cannot retrieve current position', err);
        geoLocationStatus.value = GeoLocationStatus.Error;

        if (forceUpdate) {
            snackbar.value?.showMessage('Unable to retrieve current position');
        }
    });

    geoLocationStatus.value = GeoLocationStatus.Getting;
}

function updateSpecifiedGeoLocation(coordinate: Coordinate): void {
    if (isSupportGetGeoLocationByClick() && setGeoLocationByClickMap.value) {
        transaction.value.setLatitudeAndLongitude(coordinate.latitude, coordinate.longitude);
        map.value?.setMarkerPosition(transaction.value.geoLocation);
    }
}

function clearGeoLocation(): void {
    geoMenuState.value = false;
    geoLocationStatus.value = null;
    transaction.value.removeGeoLocation();
}

function showOpenPictureDialog(): void {
    if (!canAddTransactionPicture.value || submitting.value) {
        return;
    }

    pictureInput.value?.click();
}

function uploadPicture(event: Event): void {
    if (!event || !event.target) {
        return;
    }

    const el = event.target as HTMLInputElement;

    if (!el.files || !el.files.length || !el.files[0]) {
        return;
    }

    const pictureFile = el.files[0] as File;

    el.value = '';

    uploadingPicture.value = true;
    submitting.value = true;

    transactionsStore.uploadTransactionPicture({ pictureFile }).then(response => {
        transaction.value.addPicture(response);
        uploadingPicture.value = false;
        submitting.value = false;
    }).catch(error => {
        uploadingPicture.value = false;
        submitting.value = false;

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function viewOrRemovePicture(pictureInfo: TransactionPictureInfoBasicResponse): void {
    if (mode.value !== TransactionEditPageMode.Add && mode.value !== TransactionEditPageMode.Edit) {
        window.open(getTransactionPictureUrl(pictureInfo), '_blank');
        return;
    }

    confirmDialog.value?.open('Are you sure you want to remove this transaction picture?').then(() => {
        removingPictureId.value = pictureInfo.pictureId;
        submitting.value = true;

        transactionsStore.removeUnusedTransactionPicture({ pictureInfo }).then(response => {
            if (response) {
                transaction.value.removePicture(pictureInfo);
            }

            removingPictureId.value = '';
            submitting.value = false;
        }).catch(error => {
            if (error.error && error.error.errorCode === KnownErrorCode.TransactionPictureNotFound) {
                transaction.value.removePicture(pictureInfo);
            } else if (!error.processed) {
                snackbar.value?.showError(error);
            }

            removingPictureId.value = '';
            submitting.value = false;
        });
    });
}

function onSavingTag(state: boolean): void {
    submitting.value = state;
}

function onShowDateTimeError(error: string): void {
    snackbar.value?.showError(error);
}

watch(activeTab, (newValue) => {
    if (newValue === 'map') {
        nextTick(() => {
            map.value?.initMapView();
        });
    }
});

defineExpose({
    open
});
</script>

<style>
.transaction-edit-amount .v-field__prepend-inner,
.transaction-edit-amount .v-field__append-inner,
.transaction-edit-amount .v-field__field > input {
    font-size: 1.25rem;
}

.transaction-edit-timezone.v-input input::placeholder {
    color: rgba(var(--v-theme-on-background), var(--v-high-emphasis-opacity)) !important;
    opacity: unset;
}

.transaction-edit-map-view {
    height: 220px;
}

@media (min-height: 630px) {
    .transaction-edit-map-view {
        height: 390px;
    }

    @media (min-width: 960px) {
        .transaction-pictures {
            min-height: 414px;
        }
    }
}

@media (min-height: 700px) {
    .transaction-edit-map-view {
        height: 460px;
    }

    @media (min-width: 960px) {
        .transaction-pictures {
            min-height: 484px;
        }
    }
}

@media (min-height: 780px) {
    .transaction-edit-map-view {
        height: 538px;
    }

    @media (min-width: 960px) {
        .transaction-pictures {
            min-height: 562px;
        }
    }
}

.transaction-picture .picture-control-icon {
    display: none;
    position: absolute;
    width: 100% !important;
    height: 100% !important;
    background-color: rgba(0, 0, 0, 0.4);
}

.transaction-picture .picture-control-icon > i.v-icon {
    background-color: transparent;
    color: rgba(255, 255, 255, 0.8);
}

.transaction-picture:hover .picture-control-icon,
.transaction-picture .picture-control-icon.show-control-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    vertical-align: middle;
}

.transaction-picture:hover .transaction-picture-placeholder {
    display: none;
}

.transaction-picture-add {
    border: 2px dashed rgba(var(--v-theme-grey-500));

    .transaction-picture-add-icon {
        color: rgba(var(--v-theme-grey-500));
    }
}

.transaction-picture-add.enabled:hover {
    border: 2px dashed rgba(var(--v-theme-grey-700));

    .transaction-picture-add-icon {
        color: rgba(var(--v-theme-grey-700));
    }
}
</style>
