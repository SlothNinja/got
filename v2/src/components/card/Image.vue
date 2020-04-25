<template>
  <div class='got-card' :class="show && kind ? kind : 'card-back'" ></div>
</template>

<script>
  export default {
    name: 'sn-card-image',
    props: [ 'kind', 'show' ]
  }
</script>

<style scoped lang="scss">

  @mixin got-card($class, $left) {
    .got-card.#{$class} {
      background-image:url('../../assets/cards.png');
      background-position:$left 0px;
      border-radius:10px;
      width:100px;
      height:136px;
      margin:2px;
      top:0px;
      left:0px;
    }
  }

  @include got-card(camel, 0px);
  @include got-card(carpet, -100px);
  @include got-card(coins, -200px);
  @include got-card(guard, -300px);
  @include got-card(jewels, -400px);
  @include got-card(lamp, -500px);
  @include got-card(start-camel, -600px);
  @include got-card(start-lamp, -700px);
  @include got-card(sword, -800px);
  @include got-card(turban, -900px);
  @include got-card(card-back, -1000px);

</style>
