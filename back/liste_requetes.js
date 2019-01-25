//1 •	Choisir un indicateur et un pays et voir l’évolution des valeurs - Fait
match1 = { $match: { id: 1 } };
unwind = { $unwind: "$regions" };
match2 = { $match: { "regions.region_name": "Europe & Central Asia" } }; //Il faut avoir récupéré la région avant
unwind2 = { $unwind: "$regions.countries" };
match3 = { $match: { "regions.countries.country_code": "FRA" } };
project = {
  $project: {
    "regions.countries.country_code": 1,
    "regions.countries.values": 1
  }
};
db.indicators.aggregate([match1, unwind, match2, unwind2, match3, project]);

//2 •	Choisir un indicateur et un IncomeGroup et voir les valeurs par pays - fait
match1 = { $match: { id: 1 } };
unwind = { $unwind: "$regions" };
unwind2 = { $unwind: "$regions.countries" };
match3 = { $match: { "regions.countries.income_group": "High income: OECD" } };
project = {
  $project: {
    "regions.countries.country_name": 1,
    "regions.countries.values": 1
  }
};
db.indicators.aggregate([match1, unwind, unwind2, match3, project]);

//3 •	Pour un indicateur, voir la description et la source des données - fait
db.getCollection("indicators").find(
  { indicator_code: "AG.LND.EL5M.ZS" },
  { indicator_name: 1, source_note: 1, source_organization: 1 }
);

//4 •	Voir le pays qui a le plus perdu/gagné depuis une année choisie pour l’indicateur - Fait
mapFunction = function() {
  year = 2000;
  maxYear = 2010;
  countriesValues = [];
  for (let region of this.regions) {
    for (let country of region.countries) {
      let v1 = 0;
      let v2 = 0;
      for (let value of country.values) {
        if (value.year === year) v1 = value.indicator_value;
        else if (value.year === maxYear) v2 = value.indicator_value;
      }
      if (
        v1 === Infinity ||
        v1 === -Infinity ||
        !v1 ||
        !v2 ||
        v2 === Infinity ||
        v2 === -Infinity
      )
        continue;
      countriesValues.push({
        code: country.country_code,
        growth: (v2 - v1) / v1
      });
    }
  }
  countriesValues.sort((a, b) => a.growth < b.growth);

  emit(countriesValues[0].code, countriesValues[0].growth);
  emit(
    countriesValues[countriesValues.length - 1].code,
    countriesValues[countriesValues.length - 1].growth
  );
};

reduceFunction = (key, values) => {};
queryParam = { query: { id: 84 }, out: { inline: 1 } };
db.indicators.mapReduce(mapFunction, reduceFunction, queryParam);

//5 •	Pour un indicateur, voir l’évolution de la moyenne mondiale depuis une année choisie - Fait
mapFunction = function() {
  annee = 1970;
  for (annee; annee < 2012; annee++) {
    moyenne = 0;
    nbPays = 247;
    for (var i = 0; i < this.regions.length; i++) {
      for (var j = 0; j < this.regions[i].countries.length; j++) {
        for (var k = 0; k < this.regions[i].countries[j].values.length; k++) {
          if (this.regions[i].countries[j].values[k].year == annee) {
            if (
              this.regions[i].countries[j].values[k].indicator_value != null
            ) {
              moyenne += Number(
                this.regions[i].countries[j].values[k].indicator_value
              );
            } else {
              nbPays -= 1;
            }
          }
        }
      }
    }
    if (nbPays != 0) {
      moyenne /= nbPays;
    }
    emit({ année: annee.toString(), moyenne: moyenne, nbPays: nbPays }, 1);
  }
};
reduceFunction = function() {};
queryParam = { query: { id: 84 }, out: { inline: 1 } };
db.indicators.mapReduce(mapFunction, reduceFunction, queryParam);

//6 •	Voir pour une région le pays le moins bon et le meilleur pour chaque indicateur pour une année.
mapFunction = function() {
  var year = 2000;
  var regionName = "Europe & Central Asia";
  var region = this.regions.find(r => r.region_name === regionName);
  var maxCode = "";
  var maxValue = -Infinity;
  var minCode = "";
  var minValue = Infinity;
  region.countries.forEach(c => {
    var countryV = c.values.find(v => v.year === year).indicator_value;
    if (countryV && countryV > maxValue) {
      maxValue = countryV;
      maxCode = c.country_code;
    }
    if (countryV && countryV < minValue) {
      minValue = countryV;
      minCode = c.country_code;
    }
  });
  if (!(maxCode === "" || maxValue === -Infinity))
    emit(this.id, {
      max: { code: maxCode, value: maxValue },
      min: { code: minCode, value: minValue }
    });
};

reduceFunction = (key, values) => {
  return Array.sum(values);
};
queryParam = { query: {}, out: { inline: 1 } };
db.indicators.mapReduce(mapFunction, reduceFunction, queryParam);

//7 •	Le nombre de pays - Fait
match1 = { $match: { id: 1 } };
unwind = { $unwind: "$regions" };
unwind2 = { $unwind: "$regions.countries" };
count = { $count: "country_name" };
db.indicators.aggregate([match1, unwind, unwind2, count]);

//8 •	Le nombre d’indicateurs - Fait
db.getCollection("indicators").count();

//9 •	Le nombre de pays de chaque IncomeGroup - Fait
db.getCollection("indicators").distinct("regions.countries.income_group"); //On récupère la liste des income_group différents

match1 = { $match: { id: 1 } };
unwind = { $unwind: "$regions" };
unwind2 = { $unwind: "$regions.countries" };
match2 = { $match: { "regions.countries.income_group": "High income: OECD" } };
count = { $count: "country_name" };
db.indicators.aggregate([match1, unwind, unwind2, match2, count]);

match1 = { $match: { id: 1 } };
unwind = { $unwind: "$regions" };
unwind2 = { $unwind: "$regions.countries" };
opGroup = { $group: { _id: "$income_group", tot: { $sum: 1 } } };
db.paris.aggregate([match1, unwind, unwind2, opGroup]);

//10 •	Les pays et les indicateurs les plus utilisés
