<template>
  <nav class="navbar navbar-expand-md navbar-light bg-faded">
        <button class="navbar-toggler navbar-toggler-right" type="button" data-toggle="collapse" data-target="#navbarSupportedContent" aria-controls="navbarSupportedContent" aria-expanded="false" aria-label="Toggle navigation">
            <span class="navbar-toggler-icon"></span>
        </button>
        <h4 class="my-auto"><a class="navbar-brand" href="#">{{ title }}</a></h4>
    </nav>
</template>

<script>
    export default {
        data() {
            return {
                title: "Hotshots"
            };
        }
    } 
</script>
