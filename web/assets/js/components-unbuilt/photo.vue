<template>
  <div>
    <div class="col-4-lg mt-3 align-items-center text-center">
      <p><a v-bind:href="'photos/' + photo_id + '/image.jpg'">
        <img v-bind:src="'photos/' + photo_id + '/thumb.jpg'" class="img-thumbnail">
      </a></p>
      <div class="d-inline-flex text-center">
        <div class="btn-group" role="group" v-bind:aria-label="'Photo ' + photo_id">
          <b-btn v-b-modal="'meta-' + photo_id" variant="primary">Metadata</b-btn>
          <button type="button" v-on:click="add_tag" class="btn btn-secondary">Tag</button>
          <button type="button" v-on:click="delete_photo" class="btn btn-danger">Delete</button>
        </div>
      </div>
    </div>
    <b-modal v-bind:id="'meta-' + photo_id" size="lg" title="Photo Metadata">
      <table class="table">
        <thead>
        <tr>
          <th scope="col" width="20%">Key</th>
          <th scope="col" width="80%">Value</th>
        </tr>
        </thead>
        <tbody>
          <tr v-for="(value, key) in mdata">
            <td>{{ key }}</td>
            <td>{{ value }}</td>
          </tr>
        </tbody>
      </table>
    </b-modal>
  </div>


</template>

<script>
  import bootbox from "bootbox"
  import bModal from 'bootstrap-vue/es/components/modal/modal'
  import bBtn from 'bootstrap-vue/es/components/button/button'

  export default {
    name: "photo",
    components: {
      "b-modal": bModal,
      "b-btn": bBtn
    },
    data(){
      return {
        mdata: {}
      }
    },
    props: [
      "photo_id",
    ],
    methods: {
      delete_photo: function () {
        let vue = this;
        bootbox.confirm("Do you want to delete this photo?", function (confirmed) {
          if (confirmed) {
            vue.$http.delete('photos/' + vue.photo_id).then(function (response) {
              if (!response.data.success) {
                bootbox.alert("Unable to delete photo.");
              } else {
                location.reload();
              }
            }, function (response) {
              bootbox.alert("Unable to delete photo.");
            })
          }
        })
      },
      get_metadata: function() {
        let vue = this
        this.$http.get('photos/' + this.photo_id + "/meta").then(function (response) {
          vue.mdata = response.data.photo
        });
      },
      add_tag: function() {
        let vue = this;
        bootbox.prompt("Enter tag", function (tag) {
          if(tag !== null){
            vue.$http.post('photos/' + vue.photo_id + "/tags/" + tag).then(
              function (response) {
                location.reload();
              }, function (response) {
                bootbox.alert("Unable to add tag: " + tag);
              }
            )
          }
        })
      }
    },
    mounted: function () {
      this.get_metadata()
    }
  }

</script>

<style scoped>

</style>
