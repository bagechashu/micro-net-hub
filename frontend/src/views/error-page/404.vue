<template>
  <div class="wrap">
    <div class="logo">
      <h1>404</h1>
      <p>{{ $t('error-page.6111jrs9h280') }}</p>
      <div class="sub">
        <p><a href="/">Back</a></p>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: "Page404"
};
</script>

<style scoped>
.wrap {
  margin: 0 auto;
  width: 1000px;
}
.logo {
  margin-top: 50px;
}
.logo h1 {
  font-size: 200px;
  color: #8f8e8c;
  text-align: center;
  margin-bottom: 1px;
  text-shadow: 1px 1px 6px #fff;
}
.logo p {
  color: rgb(228, 146, 162);
  font-size: 20px;
  margin-top: 1px;
  text-align: center;
}
.logo p span {
  color: lightgreen;
}
.sub a {
  color: white;
  background: #8f8e8c;
  text-decoration: none;
  padding: 7px 120px;
  font-size: 13px;
  font-family: arial, serif;
  font-weight: bold;
  border-radius: 3em;
  -webkit-border-radius: 3em;
  -moz-border-radius: 0.1em;
  -border-radius: 0.1em;
}
</style>
