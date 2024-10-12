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
  "celebrate their love",
  "commit to forever",
  "do the damn thing",
  "get hitched",
  "join forces",
  "jump the broom",
  "live happily ever after",
  "live their dreams",
  "lock it down",
  "make it official",
  'say "I do!"',
  "seal the deal",
  "settle down",
  "tie the knot",
]);


function updateMarriageTagline() {
  marriageTaglineIdx = (marriageTaglineIdx + 1) % marriageTaglines.length;
  taglineElement.innerHTML = marriageTaglines[marriageTaglineIdx];
}

var marriageTaglineInterval = 5000;
var marriageTaglineIdx = Math.floor(Math.random() * marriageTaglines.length);
var taglineElement;

window.onload = function() {
  taglineElement = document.getElementById("marriage-tagline");

  if (taglineElement !== null) {
    updateMarriageTagline();
    setInterval(updateMarriageTagline, marriageTaglineInterval);
  }
}
