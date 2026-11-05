<script>
import {SvgIcon} from '../svg.ts';
import {
  Chart,
  Tooltip,
  BarElement,
  LinearScale,
  TimeScale,
} from 'chart.js';
import {GET} from '../modules/fetch.ts';
import {Bar} from 'vue-chartjs';
import {
  startDaysBetween,
  firstStartDateAfterDate,
  fillEmptyStartDaysWithZeroes,
} from '../utils/time.js';
import {chartJsColors} from '../utils/color.ts';
import {sleep} from '../utils.ts';
import 'chartjs-adapter-dayjs-4/dist/chartjs-adapter-dayjs-4.esm';

const {pageData} = window.config;

Chart.defaults.color = chartJsColors.text;
Chart.defaults.borderColor = chartJsColors.border;

Chart.register(
  TimeScale,
  LinearScale,
  BarElement,
  Tooltip,
);

export default {
  components: {Bar, SvgIcon},
  props: {
    locale: {
      type: Object,
      required: true,
    },
  },
  data: () => ({
    isLoading: false,
    errorText: '',
    repoLink: pageData.repoLink || [],
    data: [],
  }),
  mounted() {
    this.fetchGraphData();
  },
  methods: {
    async fetchGraphData() {
      this.isLoading = true;
      try {
        let response;
        do {
          response = await GET(`${this.repoLink}/activity/recent-commits/data`);
          if (response.status === 202) {
            await sleep(1000); // wait for 1 second before retrying
          }
        } while (response.status === 202);
        if (response.ok) {
          const data = await response.json();
          const start = Object.values(data)[0].week;
          const end = firstStartDateAfterDate(new Date());
          const startDays = startDaysBetween(start, end);
          this.data = fillEmptyStartDaysWithZeroes(startDays, data).slice(-52);
          this.errorText = '';
        } else {
          this.errorText = response.statusText;
        }
      } catch (err) {
        this.errorText = err.message;
      } finally {
        this.isLoading = false;
      }
    },

    toGraphData(data) {
      return {
        datasets: [
          {
            data: data.map((i) => ({x: i.week, y: i.commits})),
            label: 'Commits',
            backgroundColor: chartJsColors.commits,
            borderWidth: 0,
            tension: 0.3,
          },
        ],
      };
    },

    getOptions() {
      return {
        responsive: true,
        maintainAspectRatio: false,
        animation: true,
        scales: {
          x: {
            type: 'time',
            grid: {
              display: false,
            },
            time: {
              minUnit: 'week',
            },
            ticks: {
              maxRotation: 0,
              maxTicksLimit: 52,
            },
          },
          y: {
            ticks: {
              maxTicksLimit: 6,
            },
          },
        },
      };
    },
  },
};
</script>
<template>
  <div>
    <h1>{{ locale.recentCommitsTitle }}</h1>
    <div class="tw-flex ui segment main-graph">
      <div v-if="isLoading || errorText !== ''" class="gt-tc tw-m-auto">
        <h2 v-if="isLoading">
          {{ locale.loadingTitle }}
        </h2>
        <h2 v-else-if="errorText">
          {{ locale.loadingTitleFailed }}
        </h2>
        <div v-if="isLoading" class="tw-flex tw-justify-center">
          <SvgIcon name="octicon-sync" class="tw-mr-2 job-status-rotate"/>
          {{ locale.loadingInfo }}
        </div>
        <div v-else class="text red tw-flex tw-justify-center">
          <SvgIcon name="octicon-x-circle-fill" class="tw-mr-2"/>
          {{ errorText }}
        </div>
      </div>
      <Bar
        v-memo="data" v-if="data.length !== 0"
        :data="toGraphData(data)" :options="getOptions()"
      />
    </div>
  </div>
</template>
<style scoped>
.main-graph {
  height: 250px;
}

h2 {
  text-align: center;
}
</style>
