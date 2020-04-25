<script>
  export default {
    methods: {
      tooltip: function (type) {
        switch (type) {
          case 'camel':
          case 'start-camel':
            return 'Move exactly 3 spaces in any direction. The spaces do not have to be in a straight line, but you cannot move over the same space twice.'
          case 'carpet':
            return 'Move in a straight line over at least one empty space.  Stop moving your thief on the first card after the empty space(s).'
          case 'coins':
            return 'Move one space and then draw an additional card during the draw step. Your hand size is permanently increased by 1.'
          case 'guard':
            return 'This card cannot be played and does nothing for you in your hand.'
          case 'jewels':
            return 'Move as if you played the card that was last played by an opponent.'
          case 'lamp':
          case 'start-lamp':
            return 'Move in a straight line until coming to the edge of the grid, an empty space, or another Thief.'
          case 'sword':
            return 'Move in a straight line until you come to another player\'s thief. Bump that thief to the next card and place your thief on the vacated card.'
          case 'turban':
            return 'Move two spaces. Claim the first Magic Item you pass over in addition to the card you claim in the Claim Magic Item step.'
          default:
            return ''
        }
      }
    }
  }
</script>
