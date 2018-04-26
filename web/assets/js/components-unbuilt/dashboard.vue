<template>
  <!--START CARD-->
  <div class="container">
    <h1 class="display-4 my-3">Your Photos</h1>
    <div class="lead d-flex justify-content-between">
      <p>Photos will automatically appear as they're taken and processed.</p>
      <input title="Tag" aria-label="Tag" v-model="tag" v-on:keyup="fetchAPIData"/>
    </div>
    <hr class="my-4">
    <div class="container">
      <div v-if="!is_loaded" class="row">
        <div class="mt-5 col-sm text-center spinner"></div>
      </div>
      <div v-else class="row">
        <div v-for="photo_id in photo_ids" class="col-sm">
          <photo :photo_id="photo_id"/>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
  import Photo from "./photo";

  export default {
    components: {
      Photo
    },
    data() {
      return {
        title: "Hotshots",
        is_loaded: false,
        photo_ids: [],
        tag: "",
        interval: null
      };
    },

    mounted: function () {
      this.fetchAPIData();

      this.interval = setInterval(function (){
        this.fetchAPIData()
      }.bind(this), 5000)
    },

    methods: {
      fetchAPIData: function () {
        let url;
        let limit = 100000
        if (this.tag !== "") {
          url = 'photos/ids?limit=' + limit + '&tag=' + this.tag
        } else {
          url = 'photos/ids?limit=' + limit
        }
        console.log(url);
        this.$http.get(url).then(function (response) {
          this.photo_ids = response.data.ids;
          this.is_loaded = true;
        }.bind(this));
      }
    }

  }
</script>
