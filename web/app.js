const { computed, createApp, onMounted, ref } = Vue;

const api = {
  async getRooms(minCapacity) {
    const query = new URLSearchParams({ minCapacity });
    return request(`/meeting/api/rooms?${query}`);
  },

  async getBookings(date) {
    const query = new URLSearchParams({ date });
    return request(`/meeting/api/bookings?${query}`);
  },

  async createBooking(booking) {
    return request('/meeting/api/bookings', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(booking),
    });
  },
};

async function request(url, options = {}) {
  const response = await fetch(url, options);
  const data = await response.json();

  if (!response.ok) {
    throw new Error(data.error || 'Не удалось выполнить запрос');
  }

  return data;
}

function today() {
  return new Date().toISOString().slice(0, 10);
}

function createEmptyForm() {
  return {
    roomId: '',
    title: '',
    organizer: '',
    start: '09:00',
    end: '10:00',
  };
}

function toBookingPayload(form, date) {
  return {
    roomId: Number(form.roomId),
    title: form.title,
    organizer: form.organizer,
    startsAt: `${date}T${form.start}:00Z`,
    endsAt: `${date}T${form.end}:00Z`,
  };
}

function formatTime(value) {
  return new Date(value).toLocaleTimeString('ru-RU', {
    hour: '2-digit',
    minute: '2-digit',
    timeZone: 'UTC',
  });
}

createApp({
  setup() {
    const date = ref(today());
    const minCapacity = ref(1);
    const rooms = ref([]);
    const bookings = ref([]);
    const form = ref(createEmptyForm());
    const message = ref('');
    const error = ref('');
    const isLoading = ref(false);
    const isSaving = ref(false);

    const visibleRooms = computed(() => {
      const capacity = Number(minCapacity.value);
      return rooms.value.filter((room) => room.capacity >= capacity);
    });

    async function load() {
      isLoading.value = true;
      error.value = '';

      try {
        const [loadedRooms, loadedBookings] = await Promise.all([
          api.getRooms(minCapacity.value),
          api.getBookings(date.value),
        ]);

        rooms.value = loadedRooms;
        bookings.value = loadedBookings;

        const selectedRoomExists = rooms.value.some(
          (room) => room.id === Number(form.value.roomId),
        );

        if (!selectedRoomExists) {
          form.value.roomId = rooms.value[0]?.id || '';
        }
      } catch (requestError) {
        error.value = requestError.message;
      } finally {
        isLoading.value = false;
      }
    }

    async function submitBooking() {
      message.value = '';
      error.value = '';
      isSaving.value = true;

      try {
        await api.createBooking(toBookingPayload(form.value, date.value));
        message.value = 'Бронь создана';
        form.value.title = '';
        await load();
      } catch (requestError) {
        error.value = requestError.message;
      } finally {
        isSaving.value = false;
      }
    }

    onMounted(load);

    return {
      bookings,
      date,
      error,
      form,
      isLoading,
      isSaving,
      load,
      message,
      minCapacity,
      submitBooking,
      visibleRooms,
      formatTime,
      rooms,
    };
  },

  template: `
    <main>
      <header>
        <span class="eyebrow">ROOMS / MVP</span>
        <h1>Переговорные</h1>
        <p>Понятное расписание без двойных броней.</p>
      </header>

      <section class="toolbar" aria-label="Фильтры расписания">
        <label>
          Дата
          <input type="date" v-model="date" @change="load">
        </label>
        <label>
          Минимум мест
          <input type="number" min="1" v-model="minCapacity" @change="load">
        </label>
      </section>

      <p v-if="error" class="error" role="alert">{{ error }}</p>
      <p v-if="isLoading" class="status" role="status">Загрузка...</p>

      <section class="layout">
        <div>
          <h2>Комнаты</h2>
          <div class="rooms">
            <article
              v-for="room in visibleRooms"
              :key="room.id"
              class="room"
            >
              <div>
                <strong>{{ room.name }}</strong>
                <span>{{ room.floor }} · {{ room.capacity }} мест</span>
              </div>
              <small>{{ room.equipment.join(' · ') || 'Базовая комплектация' }}</small>
            </article>
          </div>

          <h2>Расписание на {{ date }}</h2>
          <div
            v-for="booking in bookings"
            :key="booking.id"
            class="booking"
          >
            <time>
              {{ formatTime(booking.startsAt) }}–{{ formatTime(booking.endsAt) }}
            </time>
            <div>
              <strong>{{ booking.title }}</strong>
              <span>{{ booking.roomName }} · {{ booking.organizer }}</span>
            </div>
          </div>
          <p v-if="!bookings.length && !isLoading" class="empty">
            На эту дату броней нет.
          </p>
        </div>

        <form @submit.prevent="submitBooking">
          <h2>Новая бронь</h2>

          <label>
            Комната
            <select v-model="form.roomId" required>
              <option
                v-for="room in rooms"
                :key="room.id"
                :value="room.id"
              >
                {{ room.name }} · {{ room.capacity }} мест
              </option>
            </select>
          </label>

          <label>
            Название встречи
            <input
              v-model="form.title"
              required
              placeholder="Например, Планирование спринта"
            >
          </label>

          <label>
            Организатор
            <input
              v-model="form.organizer"
              required
              placeholder="Имя сотрудника"
            >
          </label>

          <div class="row">
            <label>
              Начало
              <input type="time" v-model="form.start" required>
            </label>
            <label>
              Конец
              <input type="time" v-model="form.end" required>
            </label>
          </div>

          <button :disabled="isSaving || !rooms.length">
            {{ isSaving ? 'Сохраняем...' : 'Забронировать' }}
          </button>
          <p v-if="message" class="message" role="status">{{ message }}</p>
        </form>
      </section>
    </main>
  `,
}).mount('#app');
