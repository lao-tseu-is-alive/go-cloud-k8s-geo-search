<style>
@import "ol/ol.css";
@import "ol-layerswitcher/dist/ol-layerswitcher.css";
.map {
  background-color: white;
  position: relative;
  top: 0px;
  bottom: 0px;
  width: 100%;
  height: 98%;
  min-height: 450px;
}

.tooltip {
  position: relative;
  padding: 3px;
  background: rgba(0, 0, 0, 0.5);
  color: white;
  opacity: 0.8;
  white-space: nowrap;
  font: 12pt sans-serif;
}
.mouse-coordinates {
  z-index: 245;
}

.ol-control:hover {
  background-color: rgba(0, 0, 0, 0);
}

.ol-control {
  font-size: 18px;
  button {
    background-color: rgba(245, 245, 245, 1);
    color: black;
    font-weight: normal;
    box-shadow:
      0px 3px 1px -2px rgba(0, 0, 0, 0.2),
      0px 2px 2px 0px rgba(0, 0, 0, 0.14),
      0px 1px 5px 0px rgba(0, 0, 0, 0.12);
    transition-property:
      box-shadow,
      transform,
      opacity,
      -webkit-box-shadow,
      -webkit-transform;
    border-radius: 4px;
  }

  button:hover {
    background-color: rgba(245, 245, 245, 1);
    color: black;
  }

  button:focus {
    background-color: rgba(245, 245, 245, 1);
    color: black;
  }
}
.ol-zoom {
  position: absolute;
  top: 21px;
  left: unset !important;
  right: 0.5em;
  margin-bottom: 1em;
  background-color: rgba(255, 255, 255, 0);

  .ol-zoom-in {
    height: 42px;
    width: 42px;
    min-width: 42px;
    color: rgba(0, 0, 0, 0.87);
    border-radius: 4px;
  }

  .ol-zoom-out {
    top: 21px;
    height: 42px;
    width: 42px;
    min-width: 42px;
    color: rgba(0, 0, 0, 0.87);
    border-radius: 4px;
  }
}
.layers_button {
  position: absolute;
  top: 197px;
  right: 0.5em;
  z-index: 250;
  background-color: #feffeb;
}

.layer-switcher-dialog {
  max-width: 250px;
  padding: 10px;
  ul {
    list-style: none;
  }

  li {
    padding-top: 0.5em;
    padding-left: 0.1em;
    text-indent: -1.5em;
  }

  label {
    padding-left: 10px;
    vertical-align: bottom;
  }
}
.gps_button {
  margin-right: -0.3em;
  top: 260px;
}

.ol-attribution {
  bottom: 1em;
  margin-right: 0.15em;
  font-size: 0.8em;
  position: fixed;
  background-color: rgba(255, 255, 255, 0);
}
</style>
<template>
    <div class="map" id="map" ref="myMap">
      <noscript> You need to have a browser with javascript support to see this OpenLayers map of Lausanne </noscript>
      <div ref="mapTooltip" class="tooltip"></div>
    </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from "vue"
import { getLog } from "../config"
import {  createLausanneMap, mapClickInfo, mapFeatureInfo } from "./Map"
import OlMap from "ol/Map"
import OlOverlay from "ol/Overlay"
// import LayerSwitcher from "ol-layerswitcher"
// import { isNullOrUndefined } from "../tools/utils"

const log = getLog("MapLausanneVue", 4, 2)
const myOlMap = ref<OlMap | null>(null)
let myMapOverlay: null | OlOverlay
const mapTooltip = ref<HTMLDivElement | null>(null)
const myProps = defineProps<{
  zoom?: number | undefined
}>()

//// EVENT SECTION

const emit = defineEmits(["map-click", "map-error"])

//// WATCH SECTION
watch(
  () => myProps.zoom,
  (val, oldValue) => {
    log.t(` watch myProps.zoom old: ${oldValue}, new val: ${val}`)
    if (val !== undefined) {
      if (val !== oldValue) {
        // do something
      }
    }
  }
  //  { immediate: true }
)

//// COMPUTED SECTION

//// FUNCTIONS SECTION
const initialize = async (center:number[]) => {
  log.t(" #> entering initialize...")
  myOlMap.value = await createLausanneMap("map", center, 8, "fonds_geo_osm_bdcad_couleur")
  if (myOlMap.value !== null) {

    myOlMap.value.on("click", (evt) => {
      const x = +Number(evt.coordinate[0]).toFixed(2)
      const y = +Number(evt.coordinate[1]).toFixed(2)
      const features: mapFeatureInfo[] = []
      if (myOlMap.value instanceof OlMap) {
        /*
        myOlMap.value.forEachFeatureAtPixel(evt.pixel, (feature, layer) => {
          let layerName = ""
          if (!isNullOrUndefined(layer)) {
            layerName = layer.get("name")
          }
          // on veut les tooltip seulement pour la couche myLayerName
          if (!isNullOrUndefined(layerName)) {
            if (layerName.indexOf(myLayerName) > -1) {
              const featureProps = feature.getProperties()
              if (!isNullOrUndefined(featureProps)) {
                const featureInfo: mapFeatureInfo = {
                  id: featureProps.id,
                  // @ts-expect-error it's ok
                  feature,
                  layer: layerName,
                  data: featureProps,
                }
                // log.l(`Feature id : ${feature_props.id}, info:`, info);
                features.push(featureInfo)
              } else {
                // @ts-expect-error it's ok
                features.push({ id: 0, feature, layer: layerName, data: null } as mapFeatureInfo)
              }
            }

          }
          // return feature
        })
        */
      } // end of forEachFeatureAtPixel
      if (features.length > 0) {
        features.forEach((featInfo) => {
          log.l("Feature found : ", featInfo)
        })
      }
      emit("map-click", { x, y, features } as mapClickInfo)
    })
    // const divToc = document.getElementById("divLayerSwitcher")
    // LayerSwitcher.renderPanel(myOlMap.value as OlMap, divToc, {})
    if (mapTooltip.value !== null) {
      myMapOverlay = new OlOverlay({
        element: mapTooltip.value as HTMLDivElement,
        offset: [0, -40],
        positioning: "top-center",
      })
      myOlMap.value.addOverlay(myMapOverlay)
    }
  }
}

onMounted(() => {
  log.t("mounted()")
  const placeStFrancoisM95 = [2538202, 1152364]
  initialize(placeStFrancoisM95)
})
</script>
