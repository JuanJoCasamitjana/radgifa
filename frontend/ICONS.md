# Iconos SVG Personalizados

## Cómo usar el componente Icon

```vue
<template>
  <Icon name="eye" />
  <Icon name="eye-off" />
  <Icon name="check" />
  <Icon name="x" />
</template>

<script setup>
import Icon from '../components/Icon.vue'
</script>
```

## Agregar nuevos iconos

### Método 1: Editar el componente Icon.vue

1. Abre `src/components/Icon.vue`
2. Agrega un nuevo `v-else-if` con tu icono:

```vue
<svg 
  v-else-if="name === 'mi-icono'" 
  width="20" 
  height="20" 
  viewBox="0 0 24 24" 
  fill="none" 
  stroke="currentColor" 
  stroke-width="2"
>
  <!-- Tu SVG path aquí -->
</svg>
```

### Método 2: Iconos desde archivos SVG

1. Coloca tus archivos SVG en `src/assets/icons/`
2. Crea un nuevo componente o modifica Icon.vue para cargar dinámicamente:

```vue
<!-- Ejemplo para cargar SVG externos -->
<img :src="`/src/assets/icons/${name}.svg`" alt="">
```

### Método 3: Librería de iconos (Recomendado para proyectos grandes)

Instala una librería como Heroicons o Lucide:

```bash
npm install @heroicons/vue
```

```vue
<template>
  <EyeIcon class="w-5 h-5" />
  <EyeSlashIcon class="w-5 h-5" />
</template>

<script setup>
import { EyeIcon, EyeSlashIcon } from '@heroicons/vue/24/outline'
</script>
```

## Fuentes de iconos SVG gratuitos

- [Heroicons](https://heroicons.com/) - Iconos minimalistas
- [Lucide](https://lucide.dev/) - Iconos modernos
- [Feather Icons](https://feathericons.com/) - Iconos simples
- [Tabler Icons](https://tabler-icons.io/) - Gran variedad
- [Phosphor Icons](https://phosphoricons.com/) - Iconos versátiles

## Personalización de iconos

Los iconos usan `currentColor`, así que heredan el color del texto:

```css
.mi-boton {
  color: #4f46e5; /* Los iconos serán azules */
}

.mi-boton:hover {
  color: #ef4444; /* Los iconos serán rojos al hover */
}
```