// Subtle ASCII animation used on the home page.
(function(){
  const el1 = document.getElementById('asciiHalo');
  const el2 = document.getElementById('asciiHalo2');
  if(!el1 || !el2) return;

  const frames = [
`   .  .     .   .
 .  . . .  .   .
   .   .   .  .
 .   .   .   .`,
` .   .   .   .
   .  .     .  
 .   . . .   .
   .   .   .  `,
`   .   .   .  
 .   .   .   .
   .  . . .  .
 .   .     .  `,
  ];

  let i = 0;
  const tick = () => {
    const f = frames[i % frames.length];
    el1.textContent = f;
    el2.textContent = f;
    i++;
  };

  tick();
  setInterval(tick, 240);
})();
