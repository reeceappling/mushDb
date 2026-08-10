import {Note} from "@/app/components/formSubcomponents/notes";
import {Nutrient} from "@/app/components/formSubcomponents/nutrients";
import {Sugar} from "@/app/components/formSubcomponents/sugars";
import {Additive} from "@/app/components/formSubcomponents/additives";
import {Grain} from "@/app/components/formSubcomponents/grains";
import {ACL, TestAcl} from "@/app/components/accessControlServer";
import CloseableSelector, {SelectorProps} from "@/app/components/selector";
import {JarRecipeSelector, NewJarRecipeForm} from "@/app/components/jarRecipeClient";

export function TestJarRecipeOK() {
    return new JarRecipeData({
        _id: "(JAR RECIPE ID HERE)",
        name: "(JAR RECIPE NAME)",
        grains: [{grain: "Oats", percentage: 100}],
        standard: true,
        nutrients: [
            {nutrient: "(NUTRIENT 1)", amount: 3.1, unit: "oz"},
            {nutrient: "(NUTRIENT 2)", amount: 3.1, unit: "oz"}
        ],
        sugars: [
            {type: "dextrose", amount: 2, unit: "g"},
            {type: "honey", amount: 2.5, unit: "drops"}
        ],
        additives: [
            {additive: "gypsum", amount: 1 / 8, unit: "tsp/cup oats"},
            {additive: "charcoal", amount: 1, unit: "pinch"}
        ],
        notes: [{time: Date.now(), note: "(TEST NOTE 1)"}, {time: Date.now() + 2000, note: "(TEST NOTE 2)"}],
        lastUpdated: 789,
        acl: TestAcl(),
    })
}

export interface JarRecipeData {
    _id: string // jarRecipeId
    name: string
    grains: Grain[]
    standard: boolean
    nutrients?: Nutrient[]
    sugars?: Sugar[]
    additives?: Additive[]
    notes?: Note[]
    lastUpdated: number
    acl: ACL
}
export class JarRecipeData {
    // Accept a single object containing the fields
    constructor(init?: Partial<JarRecipeData>) {
        // Dynamically map the object fields onto the class instance
        Object.assign(this, init);
    }

    public getId(): string {
        return this._id // TODO: should this be urlEncoded?
    }
    public getIdUrlEncoded(): string {
        return encodeURI(this.getId())
    }
    public entryType(): string {
        return "jarRecipe"
    }
    public description(): string {
        return `Jar ${this.standard?"standard ":""}recipe ${this.name} (${this._id}). Last updated on ${new Date(this.lastUpdated).toISOString()}` // TODO: nutes, sugars, antibiotics, additives, liquids?
    }
}

export function JarRecipeSelectorCloseable(sp: SelectorProps<JarRecipeData>) { // TODO: use
    const doSel = (val?: JarRecipeData):void=>{
        if (!val){
            return
        }
        sp.doSelect(val)
    }
    return <CloseableSelector<JarRecipeData> props={{
        allowCreation: sp.allowCreation,
        doSelect: doSel, // For selecting normally
        closeTxt: "Close Jar Recipe List",
        createTxt: "Create Jar Recipe",
        lowercase: "jar recipe",
        creatorInPage: sp.creatorInPage,
        createEndpt: "jarRecipe",
        createSelector:(selHdl: (onSelect: JarRecipeData) => void)=>{
            return <JarRecipeSelector allowCreate={sp.allowCreation} doSelect={(v)=>{
                v && selHdl(v)
            }}/>
        },
        createCreator:(selHdl: (onSelect: JarRecipeData) => void)=>{
            return <NewJarRecipeForm handlers={{onCreate: selHdl, isTopLevel: false}}/>
        },
    }}/>
}

// export const ChannelTextNewJarRecipe = "newJarRecipe" // TODO: USE THIS

