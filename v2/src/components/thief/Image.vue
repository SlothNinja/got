<template>
  <div :class="`thief thief-${value}`" ></div>
</template>

<script>
  export default {
    name: 'sn-thief-image',
    props: [ 'value' ]
  }
</script>

<style scoped lang="scss">
  .thief {
    top: 25px;
    left: 25px;
    width:47px;
    height:47px;
    background:url('../../assets/thieves.png');
    background-repeat:no-repeat;
    z-index:1;
    pointer-events:none;
    position: absolute;
  }

  .thief-green {
    background-position:0% 0%;
    background-size:cover;
  }

  .thief-black {
    background-position:33.3% 0%;
    background-size:cover;
  }

  .thief-purple {
    background-position:66.6% 0%;
    background-size:cover;
  }

  .thief-yellow {
    background-position:100% 0%;
    background-size:cover;
  }

  .thief-none {
    opacity: 0
  }

</style>
