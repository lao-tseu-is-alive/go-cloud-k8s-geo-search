<style>
html, body {
  margin: 0;
}
.app {
  width: 100em;
  height: 100em;
}

</style>

<template>
  <div class="app">
    <MapLausanne :zoom="3" @map-click="mapClickHandler"></MapLausanne>
  </div>
</template>

<script setup lang="ts">
import { onMounted} from "vue"
import MapLausanne from './components/MapLausanne.vue'
import { APP, getLog, BUILD_DATE, VERSION } from "./config"
import { mapClickInfo } from "./components/Map"

const log = getLog(APP, 4, 2)
const mapClickHandler = (clickInfo: mapClickInfo) => {
  log.t(`## entering... pos:${clickInfo.x}, ${clickInfo.y}`)
  log.t(`##features length :${clickInfo.features.length}`, clickInfo.features)
}
onMounted(() => {
  log.l(`Main App.vue ${APP}-${VERSION}, du ${BUILD_DATE}`)
})

</script>
