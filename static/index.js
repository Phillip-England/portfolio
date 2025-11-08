

function initToggleNav() {
    let bars = document.querySelector('#icon-bars')
    let x = document.querySelector('#icon-x')
    let overlay = document.querySelector('#overlay')
    let menu = document.querySelector('#menu-main')
    if (!bars || !x || !menu || !overlay) {
      return
    }
    bars.addEventListener('click', () => {
      x.classList.remove('hidden')
      x.classList.add('flex')
      bars.classList.add('hidden')
      bars.classList.remove('flex')
      overlay.classList.remove('hidden')
      overlay.classList.add('flex')
      menu.classList.remove('hidden')
      menu.classList.add('flex')
    })
    x.addEventListener('click', () => {
      x.classList.remove('flex')
      x.classList.add('hidden')
      bars.classList.add('flex')
      bars.classList.remove('hidden')
      overlay.classList.remove('flex')
      overlay.classList.add('hidden')
      menu.classList.remove('flex')
      menu.classList.add('hidden')
    })
    overlay.addEventListener('click', () => {
      x.classList.remove('flex')
      x.classList.add('hidden')
      bars.classList.add('flex')
      bars.classList.remove('hidden')
      overlay.classList.remove('flex')
      overlay.classList.add('hidden')
      menu.classList.remove('flex')
      menu.classList.add('hidden')
    })
}

initToggleNav()
