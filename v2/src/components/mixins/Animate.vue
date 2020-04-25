<script>
  const Velocity = require('velocity-animate')

  export default {
    methods: {
      moveTo: function (from, to, origin, done) {
        var fromRect = from.getBoundingClientRect()
        var originRect = origin.getBoundingClientRect()
        var toRect = to.getBoundingClientRect()
        var stopTop = fromRect.top - originRect.top
        var stopLeft = fromRect.left - originRect.left
        var startTop = toRect.top + 25 - originRect.top
        var startLeft = toRect.left + 25- originRect.left
        var midLeft = (startLeft-stopLeft)/2
        var midTop = (startTop-stopTop)/2
        console.log(`stop: left ${stopLeft} top ${stopTop}`)
        console.log(`start: left ${startLeft} top ${startTop}`)
        console.log(`mid: left ${midLeft} top ${midTop}`)
        from.style.left = `${startLeft}px`
        from.style.top = `${startTop}px`
        Velocity(from, { top: midTop, left: midLeft, scale: 2 }, { easing: 'easeInSine', duration: 500 })
        Velocity(from, { top: stopTop, left: stopLeft, scale: 1 }, { easing: 'easeOutSine', duration: 500, complete: done })
      },
      moveFrom: function (from, to, origin, done) {
        var fromRect = from.getBoundingClientRect()
        console.log(`moveFrom from.top: ${fromRect.top}`)
        console.log(`moveFrom from.left: ${fromRect.left}`)
        var originRect = origin.getBoundingClientRect()
        console.log(`moveFrom origin.top: ${originRect.top}`)
        console.log(`moveFrom origin.left: ${originRect.left}`)
        var toRect = to.getBoundingClientRect()
        console.log(`moveFrom to.top: ${toRect.top}`)
        console.log(`moveFrom to.left: ${toRect.left}`)
        var startTop = fromRect.top - originRect.top
        var startLeft = fromRect.left - originRect.left
        console.log(`moveFrom start.top: ${startTop}`)
        console.log(`moveFrom start.left: ${startLeft}`)
        var stopTop = toRect.top - originRect.top
        var stopLeft = toRect.left - originRect.left
        console.log(`moveFrom stop.top: ${stopTop}`)
        console.log(`moveFrom stop.left: ${stopLeft}`)
        var midLeft = (stopLeft-startLeft)/2
        var midTop = (stopTop-startTop)/2
        console.log(`moveFrom mid.top: ${midTop}`)
        console.log(`moveFrom mid.left: ${midLeft}`)
        from.style.left = `${startLeft}px`
        from.style.top = `${startTop}px`
        Velocity(from, { top: midTop, left: midLeft, scale: 2 }, { easing: 'easeInSine', duration: 500 })
        Velocity(from, { top: stopTop, left: stopLeft, scale: 1 }, { easing: 'easeOutSine', duration: 500, complete: done })
      }
    }
  }
</script>
