(function() {
  Vue.component('log-entry', {
    template: '#log-entry',
    props: ['entry'],
    computed: {
      show: function() {
        var filterType = this.$parent.filterType;
        var filterValue = this.$parent.filterValue;
        if (filterType == null) {
          return true;
        }
        if (filterType == "tag") {
          return filterValue == this.entry.tag;
        }
        return filterValue == this.entry.hostname;
      }
    },
    methods: {
      filterByHostname: function() {
        this.$emit('filter', 'hostname', this.entry.hostname);
      },
      filterByTag: function() {
        this.$emit('filter', 'tag', this.entry.tag);
      }
    }
  });

  var app = new Vue({
    el: '#main',
    data: {
      entries: [],
      filterType: null,
      filterValue: null
    },
    computed: {
      isFiltered: function() {
        return this.filterType !== null;
      }
    },
    mounted: function() {
      var websocket = new WebSocket('ws://localhost:10514/ws');
      websocket.onmessage = function(event) {
        this.entries.push(JSON.parse(event.data));
      }.bind(this);
    },
    methods: {
      filter: function(type, value) {
        this.filterType = type;
        this.filterValue = value;
      },
      removeFiltering: function() {
        this.filterType = null;
      }
    }
  });
}());
