//1 •	Choisir un indicateur et un pays et voir l’évolution des valeurs
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

//2 •	Choisir un indicateur et un IncomeGroup et voir les valeurs par pays
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

//3 •	Pour un indicateur, voir la description et la source des données
db.getCollection("indicators").find(
  { indicator_code: "AG.LND.EL5M.ZS" },
  { indicator_name: 1, source_note: 1, source_organization: 1 }
);

//4 •	Voir le pays qui a le plus perdu/gagné depuis une année choisie pour l’indicateur

//5 •	Pour un indicateur, voir l’évolution de la moyenne mondiale depuis une année choisie

//6 •	Voir pour une région le pays le moins bon et le meilleur pour chaque indicateur pour une année.

//7 •	Le nombre de pays
match1 = { $match: { id: 1 } };
unwind = { $unwind: "$regions" };
unwind2 = { $unwind: "$regions.countries" };
count = { $count: "country_name" };
db.indicators.aggregate([match1, unwind, unwind2, count]);

//8 •	Le nombre d’indicateurs
db.getCollection("indicators").count();

//9 •	Le nombre de pays de chaque IncomeGroup
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
