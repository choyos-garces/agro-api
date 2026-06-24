/// <reference path="../pb_data/types.d.ts" />

onRecordAfterDeleteSuccess(
  (e) => {
    const deletedLedgerId = e.record.id;

    // 1. Cascade Human Resources
    try {
      const attachedHumans = $app.findRecordsByFilter(
        "ledger_human_resources",
        `ledger_id = '${deletedLedgerId}'`,
      );

      for (let humanRecord of attachedHumans) {
        $app.delete(humanRecord);
      }
    } catch (err) {
      $app
        .logger()
        .error("Error cascading human resources", "error", err.message);
    }

    // 2. Cascade Machine/Asset Resources
    try {
      const attachedMachines = $app.findRecordsByFilter(
        "ledger_asset_usage",
        `ledger_id = '${deletedLedgerId}'`,
      );

      for (let machineRecord of attachedMachines) {
        $app.delete(machineRecord);
      }
    } catch (err) {
      $app
        .logger()
        .error("Error cascading machine resources", "error", err.message);
    }
  },
  "ledger_field_labor",
  "ledger_material_inputs",
);
