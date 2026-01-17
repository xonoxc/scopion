/*
 * this function formats a date into a relative time string
 * **/
export function formatTime(date: Date) {
   const now = new Date()
   const diff = now.getTime() - date.getTime()

   switch (true) {
      case diff < 60000:
         return `${Math.floor(diff / 1000)}s ago`
      case diff < 3600000:
         return `${Math.floor(diff / 60000)}m ago`

      default:
         return `${Math.floor(diff / 3600000)}h ago`
   }
}
