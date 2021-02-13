<template>
  <v-card :id='id'>
    <v-system-bar
      color='green'
      class='white--text'
    >
      <sn-user-btn
        :user='creator'
        size='x-small'
        :color='colorByUser(creator)'
      >
      </sn-user-btn>

      <span class='ml-1 white--text'>{{creator.name}}</span>

      <v-spacer></v-spacer>

      <v-tooltip bottom color='info'>
        <template v-slot:activator='{ on }'>
          <v-btn icon>
            <v-icon color='white' v-on='on'>help</v-icon>
          </v-btn>
        </template>
        <span>Help</span>
      </v-tooltip>
      <div>{{message.id}}</div>
    </v-system-bar>

    <v-card-text>
      <div>
      {{message.text}}
      </div>

      <v-divider></v-divider>

      <div class='caption'>
        Created: {{Date(message.createdAt)}}
      </div>
    </v-card-text>

  </v-card>
</template>

<script>
  import UserButton from '@/components/user/Button'
  import Color from '@/components/mixins/Color'

  export default {
    mixins: [ Color ],
    name: 'sn-message',
    props: [ 'message', 'id', 'game' ],
    components: {
      'sn-user-btn': UserButton
    },
    computed: {
      creator: function() {
        let self = this
        return {
          id: self.message.creatorId,
          name: self.message.creatorName,
          emailHash: self.message.creatorEmailHash,
          gravType: self.message.creatorGravType
        }
      }
    }
  }
</script>
