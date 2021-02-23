<template>
  <tr>
    <td class='text-left'>
      <slot></slot>
    </td>
    <td class='text-left'>
      <sn-user-btn :user="user" size="x-small"></sn-user-btn> {{user.name}}
    </td>
    <td class='text-center'>
      {{detailsFor(user.id, 'glo')}}
    </td>
    <td class='text-center'>
      {{detailsFor(user.id, 'projected')}}
    </td>
    <td class='text-center'>
      {{detailsFor(user.id, 'played')}}
    </td>
    <td class='text-center'>
      {{detailsFor(user.id, 'won')}}
    </td>
    <td class='text-center'>
      {{detailsFor(user.id, 'wp')}}%
    </td>
  </tr>
</template>

<script>
import UserButton from '@/components/user/Button'

const _ = require('lodash')

export default {
  name: 'sn-expanded-table-row',
  props: [ 'details', 'user' ],
  components: {
    'sn-user-btn': UserButton
  },
  methods: {
    detailsFor: function (id, stat) {
      let found = _.find(this.details, { 'id': id })
      if (found) {
        return _.get(found, stat, 0)
      }
      return 0
    },
  },
}
</script>
