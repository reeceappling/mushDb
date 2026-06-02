import {Note} from "@/app/components/formSubcomponents/notes";
import {Liquid} from "@/app/components/formSubcomponents/liquids";
import {Nutrient} from "@/app/components/formSubcomponents/nutrients";
import {Sugar} from "@/app/components/formSubcomponents/sugars";
import {Additive} from "@/app/components/formSubcomponents/additives";
import CloseableSelector, {SelectorProps} from "@/app/components/selector";
import {LcRecipeSelector, NewLcRecipeForm} from "@/app/components/lcRecipeClient";
import {ACL} from "@/app/components/accessControlServer";

export function TestLcRecipeOk() { // TODO: DELETEME // TODO: FIXME!
    return new LcRecipeData({
        _id: "(LC RECIPE ID HERE)",
        name: "(LC RECIPE NAME HERE)",
        liquids: [], //TODO: fixMe!
        nutrients: [],  //TODO: fixMe!
        standard: true,
        sugars: [],  //TODO: fixMe!
        additives: [], //TODO: fixMe!
        notes: [{time: Date.now(), note: "(TEST NOTE 1)"}, {time: Date.now() + 2000, note: "(TEST NOTE 2)"}],
        lastUpdated: 789,
    })
}

export interface LcRecipeData {
    _id: string
    name: string
    liquids: Liquid[]
    nutrients?: Nutrient[]
    standard: boolean
    sugars?: Sugar[]
    additives?: Additive[]
    notes?: Note[]
    lastUpdated: number
    acl?: ACL
}
export class LcRecipeData {
    // Accept a single object containing the fields
    constructor(init?: Partial<LcRecipeData>) {
        // Dynamically map the object fields onto the class instance
        Object.assign(this, init);
    }

    public getId(): string {
        return this._id
    }
    public getIdUrlEncoded(): string {
        return encodeURI(this.getId())
    }
    public entryType(): string {
        return "lcRecipe"
    }
}

export function LcRecipeSelectorCloseable(sp: SelectorProps<LcRecipeData>) { // TODO: use
    const doSel = (val?: LcRecipeData):void=>{
        if (!val){
            return
        }
        sp.doSelect(val)
    }
    return <CloseableSelector<LcRecipeData> props={{
        allowCreation: sp.allowCreation,
        doSelect: doSel, // For selecting normally
        closeTxt: "Close LC Recipe List",
        createTxt: "Create LC Recipe",
        lowercase: "lc recipe",
        creatorInPage: sp.creatorInPage,
        createEndpt: "lcRecipe",
        createSelector:(selHdl: (onSelect: LcRecipeData) => void)=>{
            return <LcRecipeSelector allowCreate={sp.allowCreation} doSelect={(v)=>{
                v && selHdl(v)
            }}/>
        },
        createCreator:(selHdl: (onSelect: LcRecipeData) => void)=>{
            return <NewLcRecipeForm handlers={{onCreate: selHdl, isTopLevel: false}}/>
        },
    }}/>
}
