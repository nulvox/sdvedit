'use strict';

// ── API helpers ──────────────────────────────────────────────────────────────

function call(fn, ...args) {
  const r = fn(...args);
  if (!r.ok) throw new Error(r.error);
  return r.data;
}

function getJSON(fn, ...args) {
  return JSON.parse(call(fn, ...args));
}

// ── State ────────────────────────────────────────────────────────────────────

const state = {
  filename: null,
  dirty: false,
};

function markDirty() {
  state.dirty = true;
  document.getElementById('save-btn').disabled = false;
}

// ── Tab system ───────────────────────────────────────────────────────────────

const TABS = ['player', 'skills', 'friendships', 'animals', 'buildings',
               'inventory', 'world', 'bundles', 'recipes', 'mail', 'quests'];

function showTab(id) {
  TABS.forEach(t => {
    document.getElementById('tab-' + t).classList.toggle('active', t === id);
    document.getElementById('panel-' + t).hidden = t !== id;
  });
  loadTab(id);
}

// ── Generic field editor helpers ─────────────────────────────────────────────

function field(label, id, value, type = 'text', extra = '') {
  return `<label class="field">
    <span>${esc(label)}</span>
    <input id="${id}" type="${type}" value="${esc(String(value))}" ${extra}>
  </label>`;
}

function numField(label, id, value, min, max) {
  return field(label, id, value, 'number', `min="${min}" max="${max}"`);
}

function selectField(label, id, value, options, labels) {
  const opts = options.map((o, i) =>
    `<option value="${esc(o)}" ${o === value ? 'selected' : ''}>${esc(labels ? labels[i] : o)}</option>`
  ).join('');
  return `<label class="field">
    <span>${esc(label)}</span>
    <select id="${id}">${opts}</select>
  </label>`;
}

function checkboxField(label, id, checked) {
  return `<label class="field field--checkbox">
    <input type="checkbox" id="${id}" ${checked ? 'checked' : ''}>
    <span>${esc(label)}</span>
  </label>`;
}

function esc(s) {
  return String(s)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

function val(id) { return document.getElementById(id).value; }
function checked(id) { return document.getElementById(id).checked; }
function num(id) { return parseInt(val(id), 10) || 0; }
function flt(id) { return parseFloat(val(id)) || 0; }

// ── Tab: Player ──────────────────────────────────────────────────────────────

function loadPlayer() {
  const p = getJSON(sdvedit_getPlayer);
  const panel = document.getElementById('panel-player');
  panel.innerHTML = `
    <h2>Player</h2>
    <div class="grid">
      ${field('Name', 'p-name', p.name)}
      ${field('Farm Name', 'p-farmName', p.farmName)}
      ${selectField('Gender', 'p-gender', p.gender, ['Male', 'Female', 'Undefined'])}
      ${numField('Money (G)', 'p-money', p.money, 0, 9999999)}
      ${numField('Current Health', 'p-health', p.health, 0, 200)}
      ${numField('Max Health', 'p-maxHealth', p.maxHealth, 20, 200)}
      ${numField('Current Stamina', 'p-stamina', p.stamina, 0, 508)}
      ${numField('Max Stamina', 'p-maxStamina', p.maxStamina, 270, 508)}
      ${numField('Deepest Mine Level', 'p-mineLevel', p.deepestMineLevel, 0, 120)}
    </div>
    <div class="actions">
      <button id="p-save-btn" class="btn-primary">Apply Changes</button>
    </div>`;

  document.getElementById('p-save-btn').addEventListener('click', () => {
    const updated = {
      name: val('p-name'),
      farmName: val('p-farmName'),
      gender: val('p-gender'),
      money: num('p-money'),
      health: num('p-health'),
      maxHealth: num('p-maxHealth'),
      stamina: num('p-stamina'),
      maxStamina: num('p-maxStamina'),
      deepestMineLevel: num('p-mineLevel'),
      experiencePoints: p.experiencePoints,
      skillLevels: p.skillLevels,
    };
    call(sdvedit_setPlayer, JSON.stringify(updated));
    markDirty();
    showToast('Player saved');
  });
}

// ── Tab: Skills ───────────────────────────────────────────────────────────────

const SKILL_NAMES = ['Farming', 'Fishing', 'Foraging', 'Mining', 'Combat', 'Luck'];
const MAX_EXP = [15000, 15000, 15000, 15000, 15000, 15000];

function loadSkills() {
  const p = getJSON(sdvedit_getPlayer);
  const rows = SKILL_NAMES.map((name, i) => `
    <tr>
      <td>${esc(name)}</td>
      <td>${numField('', 'sk-lvl-' + i, p.skillLevels[i], 0, 10)}</td>
      <td>${numField('', 'sk-exp-' + i, p.experiencePoints[i], 0, MAX_EXP[i])}</td>
    </tr>`).join('');

  document.getElementById('panel-skills').innerHTML = `
    <h2>Skills &amp; Experience</h2>
    <table class="data-table">
      <thead><tr><th>Skill</th><th>Level (0–10)</th><th>XP</th></tr></thead>
      <tbody>${rows}</tbody>
    </table>
    <div class="actions">
      <button id="sk-max-btn" class="btn-secondary">Max All Skills</button>
      <button id="sk-save-btn" class="btn-primary">Apply Changes</button>
    </div>`;

  document.getElementById('sk-max-btn').addEventListener('click', () => {
    SKILL_NAMES.forEach((_, i) => {
      document.getElementById('sk-lvl-' + i).value = 10;
      document.getElementById('sk-exp-' + i).value = MAX_EXP[i];
    });
  });

  document.getElementById('sk-save-btn').addEventListener('click', () => {
    const lvls = SKILL_NAMES.map((_, i) => num('sk-lvl-' + i));
    const exps = SKILL_NAMES.map((_, i) => num('sk-exp-' + i));
    const updated = Object.assign({}, p, {
      skillLevels: lvls,
      experiencePoints: exps,
    });
    call(sdvedit_setPlayer, JSON.stringify(updated));
    markDirty();
    showToast('Skills saved');
  });
}

// ── Tab: Friendships ─────────────────────────────────────────────────────────

const FRIENDSHIP_STATUSES = ['Friendly', 'Dating', 'Engaged', 'Married', 'Divorced'];
const HEARTS_MAX = 14; // 3500 pts

// Vanilla pet skins available per pet type (whichBreed index).
const PET_BREEDS = { Dog: [0, 1, 2], Cat: [0, 1, 2] };

function petBreedField(type, breed) {
  const breeds = PET_BREEDS[type] || [0];
  const value = breeds.includes(breed) ? breed : breeds[0];
  return selectField('Breed (skin)', 'pet-breed', value, breeds);
}

function loadFriendships() {
  const entries = getJSON(sdvedit_getFriendships);
  const pet = getJSON(sdvedit_getPet);

  const known = new Set(entries.map(e => e.npc));
  const addable = getJSON(sdvedit_knownNPCs).filter(n => !known.has(n)).sort();
  const addNpcHtml = addable.length ? `
    <div class="add-animal-form">
      <h3>Add NPC</h3>
      <p class="hint">Adds a villager you haven't met yet so you can set their friendship.</p>
      <div class="grid">
        ${selectField('Villager', 'fr-new-npc', addable[0], addable)}
      </div>
      <div class="actions">
        <button id="fr-add-btn" class="btn-primary">Add NPC</button>
      </div>
    </div>` : '';

  const rows = entries.map((e, i) => `
    <tr>
      <td class="npc-name">${esc(e.npc)}</td>
      <td>${numField('', 'fr-pts-' + i, e.points, 0, 3500)}</td>
      <td><span class="hearts">${heartsHtml(e.points)}</span></td>
      <td>${numField('', 'fr-gifts-' + i, e.giftsThisWeek, 0, 2)}</td>
      <td>${selectField('', 'fr-status-' + i, e.status, FRIENDSHIP_STATUSES)}</td>
      <td>${checkboxField('', 'fr-married-' + i, e.roommateMarriage)}</td>
    </tr>`).join('');

  let petHtml = '';
  if (pet) {
    petHtml = `
      <h3>Pet: ${esc(pet.name)} (${esc(pet.petType)})</h3>
      <div class="grid">
        ${field('Name', 'pet-name', pet.name)}
        ${selectField('Type', 'pet-type', pet.petType, ['Dog', 'Cat'])}
        ${selectField('Gender', 'pet-gender', pet.gender, ['Male', 'Female'])}
        ${numField('Friendship (0–1000)', 'pet-fr', pet.friendship, 0, 1000)}
        ${numField('Times Pet', 'pet-tp', pet.timesPet, 0, 9999)}
        ${petBreedField(pet.petType, pet.breed)}
      </div>
      <div class="actions">
        <button id="pet-save-btn" class="btn-primary">Apply Pet Changes</button>
      </div>`;
  } else {
    petHtml = `
      <h3>Pet</h3>
      <p class="hint">This save has no pet. Add one to set its name, type, and friendship.</p>
      <div class="grid">
        ${field('Name', 'pet-new-name', 'Rex')}
        ${selectField('Type', 'pet-new-type', 'Dog', ['Dog', 'Cat'])}
        ${petBreedField('Dog', 0)}
      </div>
      <div class="actions">
        <button id="pet-add-btn" class="btn-primary">Add Pet</button>
      </div>`;
  }

  document.getElementById('panel-friendships').innerHTML = `
    <h2>NPC Friendships</h2>
    <div class="actions">
      <button id="fr-max-btn" class="btn-secondary">Max All Friendships</button>
      <button id="fr-save-btn" class="btn-primary">Apply Changes</button>
    </div>
    <table class="data-table">
      <thead>
        <tr>
          <th>NPC</th><th>Points</th><th>Hearts</th>
          <th>Gifts/Wk</th><th>Status</th><th>Roommate</th>
        </tr>
      </thead>
      <tbody>${rows}</tbody>
    </table>
    ${addNpcHtml}
    ${petHtml}`;

  // Update hearts display when points change
  entries.forEach((_, i) => {
    document.getElementById('fr-pts-' + i).addEventListener('input', function () {
      this.closest('tr').querySelector('.hearts').innerHTML =
        heartsHtml(parseInt(this.value, 10) || 0);
    });
  });

  document.getElementById('fr-max-btn').addEventListener('click', () => {
    entries.forEach((_, i) => {
      document.getElementById('fr-pts-' + i).value = 2500;
      document.getElementById('fr-pts-' + i)
        .closest('tr').querySelector('.hearts').innerHTML = heartsHtml(2500);
    });
  });

  document.getElementById('fr-save-btn').addEventListener('click', () => {
    const updated = entries.map((e, i) => ({
      npc: e.npc,
      points: num('fr-pts-' + i),
      giftsThisWeek: num('fr-gifts-' + i),
      giftsToday: e.giftsToday,
      talkedToToday: e.talkedToToday,
      proposalRejected: e.proposalRejected,
      status: val('fr-status-' + i),
      roommateMarriage: checked('fr-married-' + i),
    }));
    call(sdvedit_setFriendships, JSON.stringify(updated));
    markDirty();
    showToast('Friendships saved');
  });

  // Rebuild the breed dropdown (id "pet-breed") to match the chosen pet type.
  function wireBreedToType(typeSelId) {
    const typeSel = document.getElementById(typeSelId);
    const breedSel = document.getElementById('pet-breed');
    if (!typeSel || !breedSel) return;
    typeSel.addEventListener('change', () => {
      const breeds = PET_BREEDS[typeSel.value] || [0];
      const cur = parseInt(breedSel.value, 10) || 0;
      const sel = breeds.includes(cur) ? cur : breeds[0];
      breedSel.innerHTML = breeds
        .map(b => `<option value="${b}" ${b === sel ? 'selected' : ''}>${b}</option>`)
        .join('');
    });
  }

  const addNpcBtn = document.getElementById('fr-add-btn');
  if (addNpcBtn) {
    addNpcBtn.addEventListener('click', () => {
      const npc = val('fr-new-npc');
      try {
        call(sdvedit_addFriendship, npc);
        markDirty();
        showToast(npc + ' added');
        loadFriendships();
      } catch (e) {
        showToast(e.message, 'error');
      }
    });
  }

  if (pet) {
    wireBreedToType('pet-type');

    document.getElementById('pet-save-btn').addEventListener('click', () => {
      const updated = {
        name: val('pet-name'),
        petType: val('pet-type'),
        breed: num('pet-breed'),
        friendship: num('pet-fr'),
        timesPet: num('pet-tp'),
        gender: val('pet-gender'),
      };
      call(sdvedit_setPet, JSON.stringify(updated));
      markDirty();
      showToast(updated.name + ' saved');
    });
  } else {
    wireBreedToType('pet-new-type');

    document.getElementById('pet-add-btn').addEventListener('click', () => {
      const name = val('pet-new-name').trim() || 'Pet';
      const type = val('pet-new-type');
      const breed = num('pet-breed');
      try {
        call(sdvedit_addPet, type, name, breed);
        markDirty();
        showToast(name + ' added');
        loadFriendships();
      } catch (e) {
        showToast(e.message, 'error');
      }
    });
  }
}

function heartsHtml(points) {
  const full = Math.min(Math.floor(points / 250), 14);
  const half = (points % 250) >= 125 ? 1 : 0;
  const empty = 14 - full - half;
  return '♥'.repeat(full) +
    (half ? '<span class="heart-half">♥</span>' : '') +
    '<span class="heart-empty">' + '♥'.repeat(empty) + '</span>';
}

// ── Tab: Animals ─────────────────────────────────────────────────────────────

const COOP_TYPES = new Set(['Coop', 'Big Coop', 'Deluxe Coop']);
const BARN_TYPES = new Set(['Barn', 'Big Barn', 'Deluxe Barn']);
const COOP_ANIMALS = new Set(['Chicken', 'Duck', 'Rabbit', 'Dinosaur', 'Void Chicken', 'Blue Chicken', 'Golden Chicken']);

function loadAnimals() {
  const animals = getJSON(sdvedit_getAnimals);
  const buildings = getJSON(sdvedit_getBuildings);
  const allTypes = getJSON(sdvedit_animalTypes).sort();

  const buildMap = {};
  buildings.forEach(b => { buildMap[b.id] = b.buildingType || b.xsiType || b.id; });

  const addForm = buildings.length ? `
    <div class="add-animal-form">
      <h3>Add Animal</h3>
      <div class="grid">
        ${selectField('Building', 'an-new-bldg', buildings[0]?.id || '',
            buildings.map(b => b.id),
            buildings.map(b => (b.buildingType || b.xsiType || 'Building') + ' (' + b.id.slice(0,8) + '…)'))}
        ${selectField('Type', 'an-new-type', allTypes[0] || '', allTypes)}
        ${field('Name', 'an-new-name', 'My Animal')}
      </div>
      <div class="actions">
        <button id="an-add-btn" class="btn-primary">Add Animal</button>
      </div>
    </div>` : '<p class="hint">No buildings found. Add a Barn or Coop in the Buildings tab first.</p>';

  // Buildings an animal could be moved to: same habitat (coop vs barn),
  // excluding the one it currently lives in.
  const moveTargets = (a) => buildings.filter(b => {
    if (b.id === a.buildingId) return false;
    return COOP_ANIMALS.has(a.type)
      ? COOP_TYPES.has(b.buildingType)
      : BARN_TYPES.has(b.buildingType);
  });

  const rows = animals.map((a, i) => {
    const targets = moveTargets(a);
    const moveCtl = targets.length ? `
      ${selectField('', 'an-move-' + i, targets[0].id,
          targets.map(b => b.id),
          targets.map(b => (b.buildingType || 'Building') + ' (' + b.id.slice(0, 8) + '…)'))}
      <button class="btn-sm an-move-btn" data-idx="${i}">Move</button>` : '';
    return `
    <tr>
      <td>${field('', 'an-nm-' + i, a.name)}</td>
      <td>${esc(a.type)}</td>
      <td>${esc(buildMap[a.buildingId] || a.buildingId)}</td>
      <td>${numField('', 'an-fr-' + i, a.friendship, 0, 1000)}</td>
      <td>${numField('', 'an-hp-' + i, a.happiness, 0, 255)}</td>
      <td>${numField('', 'an-fl-' + i, a.fullness, 0, 255)}</td>
      <td>${numField('', 'an-ag-' + i, a.age, 0, 999)}</td>
      <td>${numField('', 'an-pq-' + i, a.produceQuality, 0, 4)}</td>
      <td class="row-actions">
        <button class="btn-sm an-apply-btn" data-idx="${i}">Apply</button>
        ${moveCtl}
        <button class="btn-sm btn-danger an-del-btn" data-idx="${i}">Delete</button>
      </td>
    </tr>`;
  }).join('');

  const tableHtml = animals.length ? `
    <div class="actions">
      <button id="an-max-btn" class="btn-secondary">Max All Friendships</button>
    </div>
    <table class="data-table">
      <thead>
        <tr>
          <th>Name</th><th>Type</th><th>Building</th>
          <th>Friendship (0–1000)</th><th>Happiness (0–255)</th>
          <th>Fullness (0–255)</th><th>Age</th><th>Produce Quality</th><th></th>
        </tr>
      </thead>
      <tbody>${rows}</tbody>
    </table>` : '<p class="empty-state">No animals yet — add one below.</p>';

  document.getElementById('panel-animals').innerHTML =
    `<h2>Farm Animals</h2>${tableHtml}${addForm}`;

  if (animals.length) {
    document.getElementById('an-max-btn').addEventListener('click', () => {
      animals.forEach((_, i) => {
        document.getElementById('an-fr-' + i).value = 1000;
        document.getElementById('an-hp-' + i).value = 255;
      });
    });

    document.querySelectorAll('.an-apply-btn').forEach(btn => {
      btn.addEventListener('click', () => {
        const i = parseInt(btn.dataset.idx, 10);
        const a = animals[i];
        const newName = val('an-nm-' + i).trim() || a.name;
        call(sdvedit_setAnimalField, a.id, 'name', newName);
        call(sdvedit_setAnimalField, a.id, 'friendshipTowardFarmer', String(num('an-fr-' + i)));
        call(sdvedit_setAnimalField, a.id, 'happiness', String(num('an-hp-' + i)));
        call(sdvedit_setAnimalField, a.id, 'fullness', String(num('an-fl-' + i)));
        call(sdvedit_setAnimalField, a.id, 'age', String(num('an-ag-' + i)));
        call(sdvedit_setAnimalField, a.id, 'produceQuality', String(num('an-pq-' + i)));
        markDirty();
        showToast(newName + ' saved');
      });
    });

    document.querySelectorAll('.an-move-btn').forEach(btn => {
      btn.addEventListener('click', () => {
        const i = parseInt(btn.dataset.idx, 10);
        const a = animals[i];
        const target = val('an-move-' + i);
        try {
          call(sdvedit_moveAnimal, a.id, target);
          markDirty();
          showToast(a.name + ' moved');
          loadAnimals();
        } catch (err) {
          showToast('Error: ' + err.message, 'error');
        }
      });
    });

    document.querySelectorAll('.an-del-btn').forEach(btn => {
      btn.addEventListener('click', () => {
        const i = parseInt(btn.dataset.idx, 10);
        const a = animals[i];
        if (!confirm('Delete ' + a.name + '? This cannot be undone.')) return;
        try {
          call(sdvedit_removeAnimal, a.id);
          markDirty();
          showToast(a.name + ' deleted');
          loadAnimals();
        } catch (err) {
          showToast('Error: ' + err.message, 'error');
        }
      });
    });
  }

  if (buildings.length) {
    const bldgSelect = document.getElementById('an-new-bldg');
    const typeSelect = document.getElementById('an-new-type');
    const coopAnimals = new Set(['Chicken','Duck','Rabbit','Dinosaur','Void Chicken','Blue Chicken','Golden Chicken']);

    function updateTypeOptions() {
      const btype = buildMap[bldgSelect.value] || '';
      const isCoop = COOP_TYPES.has(btype);
      const isBarn = BARN_TYPES.has(btype);
      const prev = typeSelect.value;
      typeSelect.innerHTML = allTypes
        .filter(t => isCoop ? coopAnimals.has(t) : isBarn ? !coopAnimals.has(t) : true)
        .map(t => `<option value="${esc(t)}" ${t === prev ? 'selected' : ''}>${esc(t)}</option>`)
        .join('');
    }

    bldgSelect.addEventListener('change', updateTypeOptions);
    updateTypeOptions();

    document.getElementById('an-add-btn').addEventListener('click', () => {
      const buildingId = bldgSelect.value;
      const animalType = typeSelect.value;
      const animalName = val('an-new-name').trim() || animalType;
      try {
        call(sdvedit_addAnimal, buildingId, animalType, animalName);
        markDirty();
        showToast(animalName + ' added');
        loadAnimals();
      } catch (err) {
        showToast('Error: ' + err.message, 'error');
      }
    });
  }
}

// ── Tab: Buildings ────────────────────────────────────────────────────────────

function loadBuildings() {
  const buildings = getJSON(sdvedit_getBuildings);
  const buildingTypes = getJSON(sdvedit_buildingTypes).sort();

  const rows = buildings.map((b, i) => `
    <details class="building-card">
      <summary>${esc(b.buildingType || b.xsiType || 'Building')}
        <span class="muted">at (${b.tileX}, ${b.tileY})</span>
      </summary>
      <div class="grid">
        ${selectField('Type', 'bl-type-' + i, b.buildingType, buildingTypes)}
        ${checkboxField('Recompute size/occupants on type change', 'bl-recompute-' + i, true)}
        ${numField('Tile X', 'bl-tx-' + i, b.tileX, 0, 120)}
        ${numField('Tile Y', 'bl-ty-' + i, b.tileY, 0, 120)}
        ${numField('Days of Construction Left', 'bl-dc-' + i, b.daysOfConstruction, 0, 9)}
        ${numField('Days Until Upgrade', 'bl-du-' + i, b.daysUntilUpgrade, 0, 9)}
        ${numField('Hay Capacity', 'bl-hay-' + i, b.hayCapacity, 0, 999)}
        ${numField('Max Occupants', 'bl-max-' + i, b.maxOccupants, 0, 99)}
      </div>
      <fieldset><legend>Paint Color 1</legend>
        <div class="grid">
          ${checkboxField('Default', 'bl-c1d-' + i, b.paint.default1)}
          ${numField('Hue', 'bl-c1h-' + i, b.paint.h1, 0, 360)}
          ${numField('Saturation', 'bl-c1s-' + i, b.paint.s1, -100, 100)}
          ${numField('Lightness', 'bl-c1l-' + i, b.paint.l1, -100, 100)}
        </div>
      </fieldset>
      <fieldset><legend>Paint Color 2</legend>
        <div class="grid">
          ${checkboxField('Default', 'bl-c2d-' + i, b.paint.default2)}
          ${numField('Hue', 'bl-c2h-' + i, b.paint.h2, 0, 360)}
          ${numField('Saturation', 'bl-c2s-' + i, b.paint.s2, -100, 100)}
          ${numField('Lightness', 'bl-c2l-' + i, b.paint.l2, -100, 100)}
        </div>
      </fieldset>
      <fieldset><legend>Paint Color 3</legend>
        <div class="grid">
          ${checkboxField('Default', 'bl-c3d-' + i, b.paint.default3)}
          ${numField('Hue', 'bl-c3h-' + i, b.paint.h3, 0, 360)}
          ${numField('Saturation', 'bl-c3s-' + i, b.paint.s3, -100, 100)}
          ${numField('Lightness', 'bl-c3l-' + i, b.paint.l3, -100, 100)}
        </div>
      </fieldset>
      <div class="actions">
        <button class="btn-primary bl-apply-btn" data-idx="${i}">Apply Changes</button>
        <button class="btn-secondary bl-type-btn" data-idx="${i}">Change Type</button>
        <button class="btn-danger bl-del-btn" data-idx="${i}">Delete Building</button>
      </div>
    </details>`).join('');

  const addForm = `
    <div class="add-animal-form" id="add-building-form">
      <h3>Add Building</h3>
      <div class="grid">
        ${selectField('Type', 'bl-new-type', buildingTypes[0] || '', buildingTypes)}
        ${numField('Tile X', 'bl-new-tx', 10, 0, 120)}
        ${numField('Tile Y', 'bl-new-ty', 10, 0, 120)}
      </div>
      <div id="bl-placement-warnings" style="margin:0.5rem 0"></div>
      <div class="actions">
        <button class="btn-secondary" id="bl-suggest-btn">Find a spot</button>
        <button class="btn-primary" id="bl-add-btn">Add Building</button>
      </div>
    </div>`;

  document.getElementById('panel-buildings').innerHTML =
    `<h2>Farm Buildings</h2>${rows || '<p class="empty-state">No buildings placed yet.</p>'}${addForm}`;

  document.querySelectorAll('.bl-apply-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      const i = parseInt(btn.dataset.idx, 10);
      const b = buildings[i];
      const patches = [
        ['tileX', String(num('bl-tx-' + i))],
        ['tileY', String(num('bl-ty-' + i))],
        ['daysOfConstructionLeft', String(num('bl-dc-' + i))],
        ['daysUntilUpgrade', String(num('bl-du-' + i))],
        ['hayCapacity', String(num('bl-hay-' + i))],
        ['maxOccupants', String(num('bl-max-' + i))],
      ];
      patches.forEach(([field, value]) => {
        try { call(sdvedit_setBuildingField, b.id, field, value); } catch (e) { /* ignore missing */ }
      });

      const paintPatches = [
        ['buildingPaintColor/Color1Default/boolean', String(checked('bl-c1d-' + i))],
        ['buildingPaintColor/Color1Hue/int', String(num('bl-c1h-' + i))],
        ['buildingPaintColor/Color1Saturation/int', String(num('bl-c1s-' + i))],
        ['buildingPaintColor/Color1Lightness/int', String(num('bl-c1l-' + i))],
        ['buildingPaintColor/Color2Default/boolean', String(checked('bl-c2d-' + i))],
        ['buildingPaintColor/Color2Hue/int', String(num('bl-c2h-' + i))],
        ['buildingPaintColor/Color2Saturation/int', String(num('bl-c2s-' + i))],
        ['buildingPaintColor/Color2Lightness/int', String(num('bl-c2l-' + i))],
        ['buildingPaintColor/Color3Default/boolean', String(checked('bl-c3d-' + i))],
        ['buildingPaintColor/Color3Hue/int', String(num('bl-c3h-' + i))],
        ['buildingPaintColor/Color3Saturation/int', String(num('bl-c3s-' + i))],
        ['buildingPaintColor/Color3Lightness/int', String(num('bl-c3l-' + i))],
      ];
      paintPatches.forEach(([field, value]) => {
        try { call(sdvedit_setBuildingField, b.id, field, value); } catch (e) { /* ignore missing */ }
      });

      markDirty();
      showToast((b.buildingType || 'Building') + ' saved');
    });
  });

  document.querySelectorAll('.bl-type-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      const i = parseInt(btn.dataset.idx, 10);
      const b = buildings[i];
      const newType = val('bl-type-' + i);
      if (newType === b.buildingType) {
        showToast('Type unchanged');
        return;
      }
      const recompute = checked('bl-recompute-' + i);
      try {
        call(sdvedit_changeBuildingType, b.id, newType, recompute);
        markDirty();
        showToast('Building changed to ' + newType);
        loadBuildings();
      } catch (e) {
        showToast(e.message, 'error');
        loadBuildings(); // reset the dropdown to the unchanged type
      }
    });
  });

  document.querySelectorAll('.bl-del-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      const i = parseInt(btn.dataset.idx, 10);
      const b = buildings[i];
      const label = b.buildingType || 'building';
      if (!confirm('Delete this ' + label + '? This cannot be undone.')) return;
      try {
        call(sdvedit_removeBuilding, b.id);
        markDirty();
        showToast(label + ' deleted');
        loadBuildings();
      } catch (e) {
        showToast(e.message, 'error');
      }
    });
  });

  function runPlacementValidation() {
    const btype = document.getElementById('bl-new-type').value;
    const tx = num('bl-new-tx');
    const ty = num('bl-new-ty');
    const box = document.getElementById('bl-placement-warnings');
    if (!box) return;
    try {
      const warnings = getJSON(sdvedit_validatePlacement, btype, tx, ty);
      if (warnings.length === 0) {
        box.innerHTML = '<span class="badge badge--green">No issues detected</span>';
      } else {
        box.innerHTML = warnings.map(w => {
          const cls = w.severity === 'conflict' ? 'badge--red' : 'badge--yellow';
          return `<div class="placement-warn placement-warn--${w.severity}">${esc(w.message)}</div>`;
        }).join('');
      }
    } catch (_) { box.innerHTML = ''; }
  }

  ['bl-new-type', 'bl-new-tx', 'bl-new-ty'].forEach(id => {
    const el = document.getElementById(id);
    if (el) el.addEventListener('input', runPlacementValidation);
  });
  runPlacementValidation();

  document.getElementById('bl-suggest-btn').addEventListener('click', () => {
    const btype = document.getElementById('bl-new-type').value;
    try {
      const spot = getJSON(sdvedit_suggestPlacement, btype);
      document.getElementById('bl-new-tx').value = spot.tileX;
      document.getElementById('bl-new-ty').value = spot.tileY;
      runPlacementValidation();
    } catch (e) {
      showToast('No valid spot found: ' + e.message, 'error');
    }
  });

  document.getElementById('bl-add-btn').addEventListener('click', () => {
    const btype = document.getElementById('bl-new-type').value;
    const tx = num('bl-new-tx');
    const ty = num('bl-new-ty');
    try {
      call(sdvedit_addBuilding, btype, tx, ty);
      markDirty();
      showToast(btype + ' added', 'success');
      loadBuildings();
    } catch (e) {
      showToast(e.message, 'error');
    }
  });
}

// ── Tab: Inventory ────────────────────────────────────────────────────────────

function loadInventory() {
  const items = getJSON(sdvedit_getInventory);
  const catalog = getJSON(sdvedit_itemCatalog);

  const catalogOptions = catalog.map(d =>
    `<option value="${esc(d.id)}">${esc(d.name)} (#${esc(d.id)})</option>`
  ).join('');

  const rows = items.map((item, i) => {
    if (item.isNil) {
      return `<tr class="empty-slot">
        <td>${i}</td>
        <td>
          <select class="inv-new-id" data-idx="${i}">${catalogOptions}</select>
        </td>
        <td><input type="text" class="inv-new-name" data-idx="${i}" placeholder="(optional)" style="width:8em"></td>
        <td><input type="number" class="inv-new-stk" data-idx="${i}" value="1" min="1" max="999"></td>
        <td><input type="number" class="inv-new-qlt" data-idx="${i}" value="0" min="0" max="4"></td>
        <td>
          <button class="btn-sm inv-add-btn" data-idx="${i}">Add</button>
        </td>
      </tr>`;
    }
    return `<tr>
      <td>${i}</td>
      <td>${esc(item.name || item.itemId)}</td>
      <td>${esc(item.xsiType || '')}</td>
      <td>${numField('', 'inv-stk-' + i, item.stack, 1, 999)}</td>
      <td>${numField('', 'inv-qlt-' + i, item.quality, 0, 4)}</td>
      <td>
        <button class="btn-sm inv-apply-btn" data-idx="${i}">Apply</button>
        <button class="btn-sm inv-clear-btn" data-idx="${i}">Clear</button>
        <select class="inv-rep-id" data-idx="${i}">${catalogOptions}</select>
        <button class="btn-sm inv-replace-btn" data-idx="${i}">Replace</button>
      </td>
    </tr>`;
  }).join('');

  document.getElementById('panel-inventory').innerHTML = `
    <h2>Inventory</h2>
    <p class="hint">Quality: 0=Normal, 1=Silver, 2=Gold, 4=Iridium.
      Empty slots show a picker — choose an item, set stack/quality, click Add.</p>
    <table class="data-table">
      <thead>
        <tr><th>#</th><th>Name</th><th>Type</th><th>Stack</th><th>Quality</th><th></th></tr>
      </thead>
      <tbody>${rows}</tbody>
    </table>
    ${equipmentSectionHtml()}`;

  document.querySelectorAll('.inv-apply-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      const i = parseInt(btn.dataset.idx, 10);
      call(sdvedit_setInventoryItem, i, num('inv-stk-' + i), num('inv-qlt-' + i));
      markDirty();
      showToast('Slot ' + i + ' saved');
    });
  });

  document.querySelectorAll('.inv-replace-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      const i = parseInt(btn.dataset.idx, 10);
      const row = btn.closest('tr');
      const itemId = row.querySelector('.inv-rep-id').value;
      const stack = num('inv-stk-' + i);
      const quality = num('inv-qlt-' + i);
      try {
        call(sdvedit_replaceInventoryItem, i, itemId, '', stack, quality);
        markDirty();
        showToast('Slot ' + i + ' replaced');
        loadInventory();
      } catch (err) {
        showToast('Error: ' + err.message, 'error');
      }
    });
  });

  document.querySelectorAll('.inv-clear-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      const i = parseInt(btn.dataset.idx, 10);
      try {
        call(sdvedit_clearInventorySlot, i);
        markDirty();
        showToast('Slot ' + i + ' cleared');
        loadInventory();
      } catch (err) {
        showToast('Error: ' + err.message, 'error');
      }
    });
  });

  document.querySelectorAll('.inv-add-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      const i = parseInt(btn.dataset.idx, 10);
      const row = btn.closest('tr');
      const itemId = row.querySelector('.inv-new-id').value;
      const name = row.querySelector('.inv-new-name').value.trim();
      const stack = parseInt(row.querySelector('.inv-new-stk').value, 10) || 1;
      const quality = parseInt(row.querySelector('.inv-new-qlt').value, 10) || 0;
      try {
        call(sdvedit_addInventoryItem, i, itemId, name, stack, quality);
        markDirty();
        showToast('Slot ' + i + ' added');
        loadInventory();
      } catch (err) {
        showToast('Error: ' + err.message, 'error');
      }
    });
  });

  wireEquipment();
}

// ── Inventory: equipped gear ──────────────────────────────────────────────────

const EQUIP_LABELS = {
  hat: 'Hat', shirt: 'Shirt', pants: 'Pants',
  boots: 'Boots', leftRing: 'Left Ring', rightRing: 'Right Ring',
};

function equipmentSectionHtml() {
  const slots = getJSON(sdvedit_getEquipment);
  const cards = slots.map(s => {
    const label = EQUIP_LABELS[s.slot] || s.slot;
    if (!s.present) {
      // Clothing can be re-created from a verified template; hat/boots/ring
      // creation is not supported yet (no schema-verified node shape).
      if (s.kind === 'clothing') {
        return `<details class="building-card">
          <summary>${esc(label)} <span class="muted">(empty)</span></summary>
          <div class="grid">
            ${field('Item ID', 'eq-add-id-' + s.slot, '', 'text')}
            ${field('Name', 'eq-add-name-' + s.slot, '', 'text')}
          </div>
          <div class="actions">
            <button class="btn-primary eq-add-btn" data-slot="${s.slot}">Add ${esc(label)}</button>
          </div>
        </details>`;
      }
      return `<details class="building-card">
        <summary>${esc(label)} <span class="muted">(empty)</span></summary>
        <p class="hint">Nothing equipped. Adding a ${esc(label.toLowerCase())} from scratch
          isn't supported yet — equip one in-game, then edit it here.</p>
      </details>`;
    }

    let extra = '';
    if (s.kind === 'clothing') {
      extra = `
        <fieldset><legend>Color (RGBA)</legend>
          <div class="grid">
            ${numField('R', 'eq-r-' + s.slot, s.colorR, 0, 255)}
            ${numField('G', 'eq-g-' + s.slot, s.colorG, 0, 255)}
            ${numField('B', 'eq-b-' + s.slot, s.colorB, 0, 255)}
            ${numField('A', 'eq-a-' + s.slot, s.colorA, 0, 255)}
          </div>
        </fieldset>
        ${checkboxField('Dyeable', 'eq-dye-' + s.slot, s.dyeable)}`;
    } else if (s.kind === 'boots') {
      extra = `<div class="grid">
        ${numField('Defense Bonus', 'eq-def-' + s.slot, s.defenseBonus, 0, 99)}
        ${numField('Immunity Bonus', 'eq-imm-' + s.slot, s.immunityBonus, 0, 99)}
      </div>`;
    }

    return `<details class="building-card">
      <summary>${esc(label)}: ${esc(s.name || s.itemId)}</summary>
      <div class="grid">
        ${field('Item ID', 'eq-id-' + s.slot, s.itemId, 'text')}
        ${field('Name', 'eq-name-' + s.slot, s.name, 'text')}
      </div>
      ${extra}
      <div class="actions">
        <button class="btn-primary eq-apply-btn" data-slot="${s.slot}" data-kind="${s.kind}">Apply</button>
        <button class="btn-danger eq-clear-btn" data-slot="${s.slot}">Clear</button>
      </div>
    </details>`;
  }).join('');

  return `<h3 style="margin-top:1.5rem">Equipped Gear</h3>
    <p class="hint">Shirt and pants are always present. Hats, boots, and rings only
      appear once equipped in-game.</p>
    ${cards}`;
}

function wireEquipment() {
  document.querySelectorAll('.eq-apply-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      const slot = btn.dataset.slot;
      const kind = btn.dataset.kind;
      try {
        call(sdvedit_setEquipmentField, slot, 'itemId', val('eq-id-' + slot));
        call(sdvedit_setEquipmentField, slot, 'name', val('eq-name-' + slot));
        if (kind === 'clothing') {
          call(sdvedit_setEquipmentColor, slot,
            num('eq-r-' + slot), num('eq-g-' + slot), num('eq-b-' + slot), num('eq-a-' + slot));
          call(sdvedit_setEquipmentField, slot, 'dyeable', String(checked('eq-dye-' + slot)));
        } else if (kind === 'boots') {
          call(sdvedit_setEquipmentField, slot, 'defenseBonus', String(num('eq-def-' + slot)));
          call(sdvedit_setEquipmentField, slot, 'immunityBonus', String(num('eq-imm-' + slot)));
        }
        markDirty();
        showToast((EQUIP_LABELS[slot] || slot) + ' saved');
      } catch (e) {
        showToast(e.message, 'error');
      }
    });
  });

  document.querySelectorAll('.eq-clear-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      const slot = btn.dataset.slot;
      if (!confirm('Clear the ' + (EQUIP_LABELS[slot] || slot) + ' slot?')) return;
      try {
        call(sdvedit_clearEquipmentSlot, slot);
        markDirty();
        showToast((EQUIP_LABELS[slot] || slot) + ' cleared');
        loadInventory();
      } catch (e) {
        showToast(e.message, 'error');
      }
    });
  });

  document.querySelectorAll('.eq-add-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      const slot = btn.dataset.slot;
      try {
        call(sdvedit_addClothing, slot, val('eq-add-id-' + slot), val('eq-add-name-' + slot).trim());
        markDirty();
        showToast((EQUIP_LABELS[slot] || slot) + ' added');
        loadInventory();
      } catch (e) {
        showToast(e.message, 'error');
      }
    });
  });
}

// ── Tab: World ─────────────────────────────────────────────────────────────────

function loadWorld() {
  const ws = getJSON(sdvedit_getWorldState);

  document.getElementById('panel-world').innerHTML = `
    <h2>World State</h2>
    <div class="grid">
      ${selectField('Season', 'ws-season', ws.season, ['spring', 'summer', 'fall', 'winter'])}
      ${numField('Day of Month (1–28)', 'ws-day', ws.dayOfMonth, 1, 28)}
      ${numField('Year', 'ws-year', ws.year, 1, 999)}
      ${field('Daily Luck (−0.1 to 0.1)', 'ws-luck', ws.dailyLuck, 'number')}
      ${selectField('Weather Tomorrow', 'ws-weather', ws.weatherForTomorrow,
          ['Sun', 'Rain', 'Wind', 'Snow', 'Storm', 'Festival', 'Wedding'])}
      ${numField('Mine: Lowest Level Reached', 'ws-mine', ws.mineLowestLevel, 0, 120)}
      ${numField('Skull Cave: Lowest Level', 'ws-skull', ws.mineLowestForOrder, -1, 400)}
    </div>
    <div class="actions">
      <button id="ws-save-btn" class="btn-primary">Apply Changes</button>
    </div>`;

  document.getElementById('ws-save-btn').addEventListener('click', () => {
    const updated = {
      season: val('ws-season'),
      dayOfMonth: num('ws-day'),
      year: num('ws-year'),
      dailyLuck: flt('ws-luck'),
      weatherForTomorrow: val('ws-weather'),
      mineLowestLevel: num('ws-mine'),
      mineLowestForOrder: num('ws-skull'),
    };
    call(sdvedit_setWorldState, JSON.stringify(updated));
    markDirty();
    showToast('World state saved');
  });
}

// ── Tab: Bundles ──────────────────────────────────────────────────────────────

const ITEM_NAMES = {
  // Crops
  '24': 'Parsnip', '188': 'Green Bean', '190': 'Cauliflower', '192': 'Potato',
  '256': 'Tomato', '260': 'Hot Pepper', '258': 'Blueberry', '254': 'Melon',
  '270': 'Corn', '272': 'Eggplant', '276': 'Pumpkin', '280': 'Yam',
  // Forage
  '16': 'Wild Horseradish', '18': 'Daffodil', '20': 'Leek', '22': 'Dandelion',
  '396': 'Spice Berry', '398': 'Grape', '402': 'Sweet Pea',
  '404': 'Common Mushroom', '406': 'Wild Plum', '408': 'Hazelnut', '410': 'Blackberry',
  '412': 'Winter Root', '414': 'Crystal Fruit', '416': 'Snow Yam', '418': 'Crocus',
  // Animal
  '186': 'Large Milk', '182': 'Large Egg (Brown)', '174': 'Large Egg',
  '438': 'Large Goat Milk', '440': 'Wool', '442': 'Duck Egg',
  // Artisan
  '432': 'Truffle Oil', '428': 'Cloth', '426': 'Goat Cheese', '424': 'Cheese',
  '340': 'Honey', '344': 'Jelly', '613': 'Apple', '634': 'Apricot',
  '635': 'Orange', '636': 'Peach', '637': 'Pomegranate', '638': 'Cherry',
};

function itemName(id) {
  return ITEM_NAMES[id] || ('#' + id);
}

function loadBundles() {
  const bundles = getJSON(sdvedit_getBundles);

  // Group by room
  const rooms = {};
  bundles.forEach(b => {
    if (!rooms[b.room]) rooms[b.room] = [];
    rooms[b.room].push(b);
  });

  const html = Object.entries(rooms).map(([room, bs]) => {
    const bundleHtml = bs.map(b => {
      const needed = b.itemsNeeded >= 0 ? ` (need ${b.itemsNeeded} of ${b.items.length})` : '';
      const itemList = b.items.map(it =>
        `<li>${esc(itemName(it.itemId))} ×${it.quantity}${it.quality > 0 ? ' ★'.repeat(it.quality) : ''}</li>`
      ).join('');
      return `<div class="bundle-card">
        <strong>${esc(b.name)}${esc(needed)}</strong>
        <ul>${itemList}</ul>
        <div class="muted small">Reward: ${esc(b.rewardRaw)}</div>
      </div>`;
    }).join('');
    return `<details open><summary><h3>${esc(room)}</h3></summary>${bundleHtml}</details>`;
  }).join('');

  document.getElementById('panel-bundles').innerHTML =
    `<h2>Community Center Bundles</h2>
     <p class="hint">Bundle contents are shown for reference. Editing bundle completion requires modifying <code>netWorldState/completedBundles</code> which varies by game version — use mail flags to trigger bundle-related events instead.</p>
     ${html}`;
}

// ── Tab: Recipes ──────────────────────────────────────────────────────────────

function loadRecipes() {
  const cooking = getJSON(sdvedit_getCookingRecipes);
  const crafting = getJSON(sdvedit_getCraftingRecipes);

  function recipeRows(recipes, prefix) {
    return recipes.map((r, i) => `
      <tr>
        <td>${esc(r.name)}</td>
        <td>${numField('', prefix + '-tm-' + i, r.timesMade, 0, 9999)}</td>
      </tr>`).join('');
  }

  document.getElementById('panel-recipes').innerHTML = `
    <h2>Recipes</h2>
    <div class="two-col">
      <section>
        <h3>Cooking Recipes</h3>
        <div class="actions">
          <button id="ck-max-btn" class="btn-secondary">Learn All</button>
          <button id="ck-save-btn" class="btn-primary">Apply</button>
        </div>
        <table class="data-table">
          <thead><tr><th>Recipe</th><th>Times Made</th></tr></thead>
          <tbody>${recipeRows(cooking, 'ck')}</tbody>
        </table>
      </section>
      <section>
        <h3>Crafting Recipes</h3>
        <div class="actions">
          <button id="cr-max-btn" class="btn-secondary">Learn All</button>
          <button id="cr-save-btn" class="btn-primary">Apply</button>
        </div>
        <table class="data-table">
          <thead><tr><th>Recipe</th><th>Times Made</th></tr></thead>
          <tbody>${recipeRows(crafting, 'cr')}</tbody>
        </table>
      </section>
    </div>`;

  document.getElementById('ck-max-btn').addEventListener('click', () => {
    const added = call(sdvedit_learnAllRecipes, 'cookingRecipes');
    markDirty();
    showToast(added + ' cooking recipe' + (added === 1 ? '' : 's') + ' learned');
    loadRecipes();
  });
  document.getElementById('cr-max-btn').addEventListener('click', () => {
    const added = call(sdvedit_learnAllRecipes, 'craftingRecipes');
    markDirty();
    showToast(added + ' crafting recipe' + (added === 1 ? '' : 's') + ' learned');
    loadRecipes();
  });

  document.getElementById('ck-save-btn').addEventListener('click', () => {
    const updated = cooking.map((r, i) => ({ name: r.name, timesMade: num('ck-tm-' + i) }));
    call(sdvedit_setCookingRecipes, JSON.stringify(updated));
    markDirty();
    showToast('Cooking recipes saved');
  });

  document.getElementById('cr-save-btn').addEventListener('click', () => {
    const updated = crafting.map((r, i) => ({ name: r.name, timesMade: num('cr-tm-' + i) }));
    call(sdvedit_setCraftingRecipes, JSON.stringify(updated));
    markDirty();
    showToast('Crafting recipes saved');
  });
}

// ── Tab: Mail ─────────────────────────────────────────────────────────────────

function loadMail() {
  const flags = getJSON(sdvedit_getMail);

  const flagItems = (flags || []).map(f => `
    <li class="flag-item">
      <code>${esc(f)}</code>
      <button class="btn-sm btn-danger mail-remove-btn" data-flag="${esc(f)}">Remove</button>
    </li>`).join('');

  document.getElementById('panel-mail').innerHTML = `
    <h2>Mail Flags</h2>
    <p class="hint">Mail flags unlock events, NPC dialogue, and shop items. Add flags carefully.</p>
    <div class="actions">
      <input id="mail-new" type="text" placeholder="Flag name, e.g. ccIsComplete" style="width:300px">
      <button id="mail-add-btn" class="btn-primary">Add Flag</button>
    </div>
    <ul class="flag-list">${flagItems || '<li class="muted">No mail flags.</li>'}</ul>`;

  document.querySelectorAll('.mail-remove-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      const flag = btn.dataset.flag;
      call(sdvedit_removeMail, flag);
      markDirty();
      loadMail(); // re-render
    });
  });

  document.getElementById('mail-add-btn').addEventListener('click', () => {
    const flag = document.getElementById('mail-new').value.trim();
    if (!flag) return;
    call(sdvedit_addMail, flag);
    markDirty();
    loadMail();
  });
}

// ── Tab: Quests ───────────────────────────────────────────────────────────────

function loadQuests() {
  const quests = getJSON(sdvedit_getQuests);

  const rows = (quests || []).map(q => `
    <tr>
      <td>${q.id}</td>
      <td>${esc(q.title || '—')}</td>
      <td>${esc(q.questType || '—')}</td>
      <td>${q.moneyReward > 0 ? q.moneyReward + 'G' : '—'}</td>
      <td><span class="badge ${q.completed ? 'badge--green' : 'badge--grey'}">
        ${q.completed ? 'Done' : 'Active'}
      </span></td>
      <td>${q.daysLeft > 0 ? q.daysLeft + 'd' : '—'}</td>
    </tr>`).join('');

  document.getElementById('panel-quests').innerHTML = `
    <h2>Quest Log</h2>
    <p class="hint">Quest editing is read-only — completing quests mid-game can cause script state issues.
    Use mail flags to trigger quest-unlocking events instead.</p>
    <table class="data-table">
      <thead>
        <tr><th>ID</th><th>Title</th><th>Type</th><th>Reward</th><th>Status</th><th>Days Left</th></tr>
      </thead>
      <tbody>${rows || '<tr><td colspan="6" class="muted">No active quests.</td></tr>'}</tbody>
    </table>`;
}

// ── Tab loader dispatcher ─────────────────────────────────────────────────────

const tabLoaders = {
  player: loadPlayer,
  skills: loadSkills,
  friendships: loadFriendships,
  animals: loadAnimals,
  buildings: loadBuildings,
  inventory: loadInventory,
  world: loadWorld,
  bundles: loadBundles,
  recipes: loadRecipes,
  mail: loadMail,
  quests: loadQuests,
};

function loadTab(id) {
  try {
    tabLoaders[id]?.();
  } catch (err) {
    document.getElementById('panel-' + id).innerHTML =
      `<div class="error-box">Error loading ${id}: ${esc(err.message)}</div>`;
  }
}

// ── File I/O ──────────────────────────────────────────────────────────────────

function initFileIO() {
  const dropzone = document.getElementById('dropzone');
  const fileInput = document.getElementById('file-input');

  dropzone.addEventListener('dragover', e => { e.preventDefault(); dropzone.classList.add('drag-over'); });
  dropzone.addEventListener('dragleave', () => dropzone.classList.remove('drag-over'));
  dropzone.addEventListener('drop', e => {
    e.preventDefault();
    dropzone.classList.remove('drag-over');
    loadFile(e.dataTransfer.files[0]);
  });
  dropzone.addEventListener('click', () => fileInput.click());
  fileInput.addEventListener('change', () => loadFile(fileInput.files[0]));

  document.getElementById('save-btn').addEventListener('click', exportFile);
}

async function loadFile(file) {
  if (!file) return;
  const buf = await file.arrayBuffer();
  const arr = new Uint8Array(buf);
  try {
    call(sdvedit_load, arr);
  } catch (err) {
    showToast('Error: ' + err.message, 'error');
    return;
  }
  state.filename = file.name;
  state.dirty = false;
  document.getElementById('dropzone').hidden = true;
  document.getElementById('editor').hidden = false;
  document.getElementById('file-label').textContent = file.name;
  document.getElementById('save-btn').disabled = true;
  showTab('player');
}

function exportFile() {
  let bytes;
  try {
    const result = sdvedit_export();
    if (!result.ok) throw new Error(result.error);
    bytes = result.data;
  } catch (err) {
    showToast('Export error: ' + err.message, 'error');
    return;
  }
  const blob = new Blob([bytes], { type: 'application/octet-stream' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = state.filename || 'save_edited';
  a.click();
  URL.revokeObjectURL(url);
  state.dirty = false;
  document.getElementById('save-btn').disabled = true;
  showToast('Save file downloaded');
}

// ── Toast ─────────────────────────────────────────────────────────────────────

function showToast(msg, type = 'success') {
  const t = document.getElementById('toast');
  t.textContent = msg;
  t.className = 'toast toast--' + type + ' show';
  clearTimeout(t._timer);
  t._timer = setTimeout(() => t.classList.remove('show'), 2500);
}

// ── Init ──────────────────────────────────────────────────────────────────────

document.addEventListener('sdvedit:ready', () => {
  initFileIO();
  TABS.forEach(id => {
    document.getElementById('tab-' + id).addEventListener('click', () => showTab(id));
  });
  document.getElementById('wasm-loading').hidden = true;
  document.getElementById('dropzone').hidden = false;
});

document.addEventListener('DOMContentLoaded', () => {
  // Warn on unsaved changes
  window.addEventListener('beforeunload', e => {
    if (state.dirty) { e.preventDefault(); e.returnValue = ''; }
  });
});
