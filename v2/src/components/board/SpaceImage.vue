<template>
  <div
    class='board-space-image'
    :class='kind'
  >
  </div>
</template>

<script>
  export default {
    name: 'sn-space-image',
    props: ['value'],
    computed: {
      kind: function () {
        var self = this
        return self.value ? self.value.kind : 'none'
      }
    }
  }
</script>

<!-- Add 'scoped' attribute to limit CSS to this component only -->
<style scoped lang='scss'>
  @mixin space($class, $left, $top, $height, $width) {
    .board-space-image.#{$class} {
      background-image:url('../../assets/cards.png');
      background-position: ($left - 7px) ($top - 9px);
      border-radius:10px;
      height:$height;
      width:$width;
      position:absolute;
      top:(90px - $height) / 2;
      left:(90px - $width) / 2;
    }
    .board-space-image.#{$class}.grow-leave-active {
      transition-delay: 1s;
      z-index:100;
      transition: height .5s, width .5s, margin .5s, backgroud-position .5s;
    }
    .board-space-image.#{$class}.grow-leave-to {
      height: 136px;
      width: 100px;
      margin: 2px;
      background-position: $left $top;
    }
  }

  @include space(camel, -2px, -1px, 82px, 82px);
  @include space(carpet, -102px, -1px, 82px, 82px);
  @include space(coins, -202px, -1px, 82px, 82px);
  @include space(guard, -302px, -1px, 82px, 82px);
  @include space(jewels, -402px, -1px, 82px, 82px);
  @include space(lamp, -502px, -1px, 82px, 82px);
  @include space(sword, -802px, -1px, 82px, 82px);
  @include space(turban, -902px, -1px, 82px, 82px);

</style>
