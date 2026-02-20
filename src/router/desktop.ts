import { type NavigationGuardReturn, createRouter, createWebHashHistory } from 'vue-router';

import { TemplateType } from '@/core/template.ts';
import { isUserLogined, isUserUnlocked } from '@/lib/userstate.ts';

import MainLayout from '@/views/desktop/MainLayout.vue';
import LoginPage from '@/views/desktop/LoginPage.vue';
import SignUpPage from '@/views/desktop/SignupPage.vue';
import VerifyEmailPage from '@/views/desktop/VerifyEmailPage.vue';
import ForgetPasswordPage from '@/views/desktop/ForgetPasswordPage.vue';
import ResetPasswordPage from '@/views/desktop/ResetPasswordPage.vue';
import OAuth2CallbackPage from '@/views/desktop/OAuth2CallbackPage.vue';
import UnlockPage from '@/views/desktop/UnlockPage.vue';

import HomePage from '@/views/desktop/HomePage.vue';

import TransactionListPage from '@/views/desktop/transactions/ListPage.vue';
import CalendarPage from '@/views/desktop/calendar/CalendarPage.vue';

import StatisticsTransactionPage from '@/views/desktop/statistics/TransactionPage.vue';

import InsightsExplorerPage from '@/views/desktop/insights/ExplorerPage.vue';

import AccountListPage from '@/views/desktop/accounts/ListPage.vue';

import TransactionCategoryListPage from '@/views/desktop/categories/ListPage.vue';

import TransactionTagListPage from '@/views/desktop/tags/ListPage.vue';

import CounterpartyListPage from '@/views/desktop/counterparties/ListPage.vue';

import CFOListPage from '@/views/desktop/cfos/ListPage.vue';

import LocationListPage from '@/views/desktop/locations/ListPage.vue';

import AssetListPage from '@/views/desktop/assets/ListPage.vue';

import InvestorListPage from '@/views/desktop/investors/ListPage.vue';

import BudgetListPage from '@/views/desktop/budgets/ListPage.vue';

import ObligationListPage from '@/views/desktop/obligations/ListPage.vue';

import TaxRecordListPage from '@/views/desktop/taxes/ListPage.vue';

import CashFlowPage from '@/views/desktop/reports/CashFlowPage.vue';
import PnLPage from '@/views/desktop/reports/PnLPage.vue';
import BalancePage from '@/views/desktop/reports/BalancePage.vue';
import PaymentCalendarPage from '@/views/desktop/reports/PaymentCalendarPage.vue';

import TransactionTemplateListPage from '@/views/desktop/templates/ListPage.vue';

import UserSettingsPage from '@/views/desktop/user/UserSettingsPage.vue';
import AppSettingsPage from '@/views/desktop/app/AppSettingsPage.vue';

import ExchangeRatesListPage from '@/views/desktop/exchangerates/ListPage.vue';
import AboutPage from '@/views/desktop/AboutPage.vue';

function checkLogin(): NavigationGuardReturn {
    if (!isUserLogined()) {
        return {
            path: '/login',
            replace: true
        };
    }

    if (!isUserUnlocked()) {
        return {
            path: '/unlock',
            replace: true
        };
    }

    return true;
}

function checkLocked(): NavigationGuardReturn {
    if (!isUserLogined()) {
        return {
            path: '/login',
            replace: true
        };
    }

    if (isUserUnlocked()) {
        return {
            path: '/',
            replace: true
        };
    }

    return true;
}

function checkNotLogin(): NavigationGuardReturn {
    if (isUserLogined() && !isUserUnlocked()) {
        return {
            path: '/unlock',
            replace: true
        };
    }

    if (isUserLogined()) {
        return {
            path: '/',
            replace: true
        };
    }

    return true;
}

const router = createRouter({
    history: createWebHashHistory(),
    routes: [
        {
            path: '/',
            component: MainLayout,
            beforeEnter: checkLogin,
            children: [
                {
                    path: '',
                    component: HomePage,
                    beforeEnter: checkLogin
                },
                {
                    path: '/transaction/list',
                    component: TransactionListPage,
                    beforeEnter: checkLogin,
                    props: route => ({
                        initPageType: route.query['pageType'],
                        initDateType: route.query['dateType'],
                        initMaxTime: route.query['maxTime'],
                        initMinTime: route.query['minTime'],
                        initType: route.query['type'],
                        initCategoryIds: route.query['categoryIds'],
                        initAccountIds: route.query['accountIds'],
                        initCounterpartyId: route.query['counterpartyId'],
                        initTagFilter: route.query['tagFilter'],
                        initAmountFilter: route.query['amountFilter'],
                        initKeyword: route.query['keyword']
                    })
                },
                {
                    path: '/calendar',
                    component: CalendarPage,
                    beforeEnter: checkLogin
                },
                {
                    path: '/statistics/transaction',
                    component: StatisticsTransactionPage,
                    beforeEnter: checkLogin,
                    props: route => ({
                        initAnalysisType: route.query['analysisType'],
                        initChartDataType: route.query['chartDataType'],
                        initChartType: route.query['chartType'],
                        initChartDateType: route.query['chartDateType'],
                        initStartTime: route.query['startTime'],
                        initEndTime: route.query['endTime'],
                        initFilterAccountIds: route.query['filterAccountIds'],
                        initFilterCategoryIds: route.query['filterCategoryIds'],
                        initTagFilter: route.query['tagFilter'],
                        initKeyword: route.query['keyword'],
                        initSortingType: route.query['sortingType'],
                        initTrendDateAggregationType: route.query['trendDateAggregationType'],
                        initAssetTrendsDateAggregationType: route.query['assetTrendsDateAggregationType']
                    })
                },
                {
                    path: '/insights/explorer',
                    component: InsightsExplorerPage,
                    beforeEnter: checkLogin,
                    props: route => ({
                        initId: route.query['id'],
                        initActiveTab: route.query['activeTab'],
                        initDateRangeType: route.query['dateRangeType'],
                        initStartTime: route.query['startTime'],
                        initEndTime: route.query['endTime']
                    })
                },
                {
                    path: '/account/list',
                    component: AccountListPage,
                    beforeEnter: checkLogin
                },
                {
                    path: '/category/list',
                    component: TransactionCategoryListPage,
                    beforeEnter: checkLogin
                },
                {
                    path: '/tag/list',
                    component: TransactionTagListPage,
                    beforeEnter: checkLogin
                },
                {
                    path: '/counterparty/list',
                    component: CounterpartyListPage,
                    beforeEnter: checkLogin
                },
                {
                    path: '/cfo/list',
                    component: CFOListPage,
                    beforeEnter: checkLogin
                },
                {
                    path: '/location/list',
                    component: LocationListPage,
                    beforeEnter: checkLogin
                },
                {
                    path: '/asset/list',
                    component: AssetListPage,
                    beforeEnter: checkLogin
                },
                {
                    path: '/investor/list',
                    component: InvestorListPage,
                    beforeEnter: checkLogin
                },
                {
                    path: '/budget/list',
                    component: BudgetListPage,
                    beforeEnter: checkLogin
                },
                {
                    path: '/obligation/list',
                    component: ObligationListPage,
                    beforeEnter: checkLogin
                },
                {
                    path: '/tax/list',
                    component: TaxRecordListPage,
                    beforeEnter: checkLogin
                },
                {
                    path: '/report/cashflow',
                    component: CashFlowPage,
                    beforeEnter: checkLogin
                },
                {
                    path: '/report/pnl',
                    component: PnLPage,
                    beforeEnter: checkLogin
                },
                {
                    path: '/report/balance',
                    component: BalancePage,
                    beforeEnter: checkLogin
                },
                {
                    path: '/report/payment-calendar',
                    component: PaymentCalendarPage,
                    beforeEnter: checkLogin
                },
                {
                    path: '/template/list',
                    component: TransactionTemplateListPage,
                    beforeEnter: checkLogin,
                    props: {
                        initType: TemplateType.Normal.type
                    }
                },
                {
                    path: '/schedule/list',
                    component: TransactionTemplateListPage,
                    beforeEnter: checkLogin,
                    props: {
                        initType: TemplateType.Schedule.type
                    }
                },
                {
                    path: '/exchange_rates',
                    component: ExchangeRatesListPage,
                    beforeEnter: checkLogin
                },
                {
                    path: '/user/settings',
                    component: UserSettingsPage,
                    beforeEnter: checkLogin,
                    props: route => ({
                        initTab: route.query['tab']
                    })
                },
                {
                    path: '/app/settings',
                    component: AppSettingsPage,
                    beforeEnter: checkLogin,
                    props: route => ({
                        initTab: route.query['tab']
                    })
                },
                {
                    path: '/about',
                    component: AboutPage,
                    beforeEnter: checkLogin
                }
            ]
        },
        {
            path: '/login',
            component: LoginPage,
            beforeEnter: checkNotLogin
        },
        {
            path: '/signup',
            component: SignUpPage,
            beforeEnter: checkNotLogin
        },
        {
            path: '/verify_email',
            component: VerifyEmailPage,
            props: route => ({
                email: route.query['email'],
                token: route.query['token'],
                hasValidEmailVerifyToken: route.query['emailSent'] === 'true'
            })
        },
        {
            path: '/forgetpassword',
            component: ForgetPasswordPage,
            beforeEnter: checkNotLogin
        },
        {
            path: '/resetpassword',
            component: ResetPasswordPage,
            props: route => ({
                token: route.query['token']
            })
        },
        {
            path: '/oauth2_callback',
            component: OAuth2CallbackPage,
            props: route => ({
                token: route.query['token'],
                provider: route.query['provider'],
                platform: route.query['platform'],
                userName: route.query['userName'],
                errorCode: route.query['errorCode'],
                errorMessage: route.query['errorMessage']
            })
        },
        {
            path: '/unlock',
            component: UnlockPage,
            beforeEnter: checkLocked
        }
    ],
})

export default router;
