// Thanks, Stack Overflow.
// https://stackoverflow.com/a/6274381
function shuffle(a) {
    var j, x, i;
    for (i = a.length - 1; i > 0; i--) {
        j = Math.floor(Math.random() * (i + 1));
        x = a[i];
        a[i] = a[j];
        a[j] = x;
    }
    return a;
}

marriageTaglines = shuffle([
  "celebrated their love",
  "committed to forever",
  "did the damn thing",
  "exchanged vows",
  "got hitched",
  "joined forces",
  "jumped the broom",
  "lit up the dance floor",
  "lived happily ever after",
  "locked it down",
  "made it official",
  'said "I do!"',
  "sealed the deal",
  "settled down",
  "took on the world",
  "tied the knot",
]);


function updateMarriageTagline() {
  marriageTaglineIdx = (marriageTaglineIdx + 1) % marriageTaglines.length;
  taglineElement.innerHTML = marriageTaglines[marriageTaglineIdx];
}

var marriageTaglineInterval = 2500;
var marriageTaglineIdx = Math.floor(Math.random() * marriageTaglines.length);
var taglineElement;

window.onload = function() {
  taglineElement = document.getElementById("marriage-tagline");

  if (taglineElement !== null) {
    updateMarriageTagline();
    setInterval(updateMarriageTagline, marriageTaglineInterval);
  }
}
