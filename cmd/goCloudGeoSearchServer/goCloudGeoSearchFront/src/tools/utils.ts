/*eslint @typescript-eslint/no-explicit-any: "off"*/
import { getLog } from "../config"

const log = getLog("Utils", 2, 2)
/**
 * check if the given variable is null or undefined
 * @param variable
 * @returns true if given variable is null or undefined, false in other cases
 */
export const isNullOrUndefined = (variable: any): boolean => typeof variable === "undefined" || variable === null

/**
 * check if the given variable is null, undefined or an empty string(zero length)
 * @param variable
 * @returns true if given variable is null or undefined or an empty string, false in other cases
 */
export const isEmpty = (variable: any): boolean =>
  typeof variable === "undefined" || variable === null || variable === ""

/**
 * convert a date string from iso yyyy-mm-dd in french europe dd-mm-yyyy
 * @param strIsoDate string yyyy-mm-dd
 * @returns date as string in format dd-mm-yyyy
 */
export const dateIso2Fr = function (strIsoDate: string): string {
  if (isEmpty(strIsoDate)) {
    return ""
  }
  const [y, m, d] = strIsoDate.split("-")
  return [d, m, y].join("-")
}

export const getDateFromTimeStamp = (isoDate: string) => {
  log.t(`#> entering : ${isoDate}`, isoDate)
  if (typeof isoDate !== "string") return "not_date"
  if (isNullOrUndefined(isoDate) || isoDate.indexOf("T") < 0) return ""
  const dateTS = dateIso2Fr(isoDate.split("T")[0])
  log.l(`dateTS : ${dateTS}`, isoDate)
  return dateTS
}

export const getDateIsoFromTimeStamp = (isoDate: string): string => {
  log.t(`#> entering : ${isoDate}`, isoDate)
  if (typeof isoDate !== "string") return "not_date"
  if (isNullOrUndefined(isoDate) || isoDate.indexOf("T") < 0) return ""
  const dateISO = isoDate.split("T")[0]
  //const dateISO = `${dateTS.getFullYear()}-${dateTS.getMonth()}-${dateTS.getDate()}`
  log.l(`dateTS : ${dateISO}`)
  return dateISO
}

export const truncateText = (text: string, maxSize = 40): string => {
  if (isNullOrUndefined(text)) return ""
  if (text.length < maxSize) return text
  return `${text.substring(0, maxSize)}…`
}

export const parseJsonWithDetailedError = (jsonString: string, context: number = 50): any => {
  log.t(">parseJsonWithDetailedError ")
  try {
    // Attempt to parse the JSON string
    const result = JSON.parse(jsonString)
    log.l("Parsing successful", result)
    return result
  } catch (error) {
    if (error instanceof SyntaxError) {
      // Extracting approximate position information from the error message
      const match = error.message.match(/position (\d+)/)
      if (match) {
        const position = parseInt(match[1], 10)
        // Calculating line and column based on the position
        const lines = jsonString.substring(0, position).split("\n")
        const line = lines.length
        const column = lines[lines.length - 1].length + 1
        // Extracting a 30-character excerpt around the error position
        const start = Math.max(0, position - context)
        const end = Math.min(jsonString.length, position + context)
        const excerpt = `...${jsonString.slice(start, position)}«${jsonString.charAt(position)}»${jsonString.substring(position + 1, end)}...`

        log.w(`Error parsing JSON at [${line}:${column}] ${error.message}: "${excerpt}"`)
        log.l(jsonString)
      } else {
        log.w("Error parsing JSON:", error.message)
      }
    } else {
      // Non-syntax errors (unlikely in this context, but good practice to handle)
      log.w("Unexpected error:", error)
    }
  }
}
