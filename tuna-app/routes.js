//SPDX-License-Identifier: Apache-2.0
import { Router } from "express";
import tuna from "./controller.js";

const router = new Router();

router.get("/get_tuna/:id", function(req, res) {
  tuna.get_tuna(req, res);
});
router.get("/add_tuna/:tuna", function(req, res) {
  tuna.add_tuna(req, res);
});
router.get("/get_all_tuna", function(req, res) {
  tuna.get_all_tuna(req, res);
});
router.get("/change_holder/:holder", function(req, res) {
  tuna.change_holder(req, res);
});

export default router;
